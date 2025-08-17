package main

import(
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/neriAle/chirpy/internal/auth"
	"github.com/neriAle/chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateUser(rw http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email 		string `json:"email"`
		Password 	string `json:"password"`
	}
	type userParameters struct {
		Email          string
		HashedPassword string
	}
	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(rw, 400, "Email and Password are required for registration")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing the password: %s", err)
		respondWithError(rw, 500, "Something went wrong while hashing the password")
		return
	}

	userParams := userParameters{Email: params.Email, HashedPassword: hash}
	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams(userParams))
	if err != nil {
		log.Printf("Error creating the user on the database: %s", err)
		respondWithError(rw, 500, "Can't create user")
		return
	}

	mappedUser := User(user)
	respondWithJSON(rw, 201, mappedUser)
}

func (cfg *apiConfig) handlerUpdateUser(rw http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Header is missing JWT: %s", err)
		respondWithError(rw, 401, "Header is missing JWT")
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("Invalid token: %s", err)
		respondWithError(rw, 401, "JWT is not valid")
		return
	}

	type parameters struct {
		Email 		string `json:"email"`
		Password 	string `json:"password"`
	}
	type updateUserParameters struct {
		Email          string
		HashedPassword string
		ID             uuid.UUID
	}
	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(rw, 400, "Email and Password are required for updating")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing the password: %s", err)
		respondWithError(rw, 500, "Something went wrong while hashing the password")
		return
	}

	updateParams := updateUserParameters{Email: params.Email, HashedPassword: hash, ID: userId}
	user, err := cfg.db.UpdateUser(req.Context(), database.UpdateUserParams(updateParams))
	if err != nil {
		log.Printf("Error updating the user on the database: %s", err)
		respondWithError(rw, 500, "Can't update user")
		return
	}

	mappedUser := User(user)
	respondWithJSON(rw, 200, mappedUser)
}

func (cfg *apiConfig) handlerLoginUser(rw http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email 				string `json:"email"`
		Password 			string `json:"password"`
	}
	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(rw, 400, "Email and Password are required for login")
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		log.Printf("User not found: %s", err)
		respondWithError(rw, 401, "Incorrect email or password")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Invalid password: %s", err)
		respondWithError(rw, 401, "Incorrect email or password")
		return
	}

	expiration := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, expiration)
	if err != nil {
		log.Printf("Couldn't sign the JWT: %s", err)
		respondWithError(rw, 500, "Unable to sign the JWT")
		return
	}

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Couldn't create refresh token: %s", err)
		respondWithError(rw, 500, "Unable to create refresh token")
		return
	}

	type refreshTokenParams struct {
		Token  		string
		ExpiresAt 	time.Time
		UserID 		uuid.UUID
	}
	refTokPar := refreshTokenParams{
		Token:		refresh_token,
		ExpiresAt:	time.Now().AddDate(0, 0, 60),
		UserID:		user.ID,
	}
	refresh_token, err = cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams(refTokPar))
	if err != nil {
		log.Printf("Couldn't store refresh token: %s", err)
		respondWithError(rw, 500, "Unable to store refresh token")
		return
	}

	type UserWithToken struct {
		User
		Token			string	`json:"token,omitempty"`
		Refresh_token	string	`json:"refresh_token"`
	}

	usr := UserWithToken{
		User: User{
			ID: user.ID, 
			CreatedAt: user.CreatedAt, 
			UpdatedAt: user.UpdatedAt, 
			Email: user.Email,
		}, 
		Token: token,
		Refresh_token: refresh_token,
	}
	respondWithJSON(rw, 200, usr)
}

func (cfg *apiConfig) handlerRefreshJWT(rw http.ResponseWriter, req *http.Request) {
	refresh_token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Header is missing refresh token: %s", err)
		respondWithError(rw, 401, "Header is missing refresh token")
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), refresh_token)
	if err != nil {
		log.Printf("Refresh token expired: %s", err)
		respondWithError(rw, 401, "Refresh token expired")
		return
	}

	expiration := time.Hour
	token, err := auth.MakeJWT(user, cfg.tokenSecret, expiration)
	if err != nil {
		log.Printf("Couldn't sign the JWT: %s", err)
		respondWithError(rw, 500, "Unable to sign the JWT")
		return
	}

	type token_return struct {
		Token	string	`json:"token"`
	}
	respondWithJSON(rw, 200, token_return{Token: token})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(rw http.ResponseWriter, req *http.Request) {
	refresh_token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Header is missing refresh token: %s", err)
		respondWithError(rw, 400, "Header is missing refresh token")
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), refresh_token)
	if err != nil {
		log.Printf("Unable to revoke token: %s", err)
		respondWithError(rw, 404, "Refresh token not found")
		return
	}

	rw.WriteHeader(204)
}