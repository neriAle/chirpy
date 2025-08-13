package main

import(
	"github.com/google/uuid"
	"github.com/neriAle/chirpy/internal/auth"
	"github.com/neriAle/chirpy/internal/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileserverHits 	atomic.Int32
	db 				*database.Queries
	platform 		string
	tokenSecret 	string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string	`json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(rw, req)
	})
}

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
	return
}

func (cfg *apiConfig) handlerLoginUser(rw http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email 		string `json:"email"`
		Password 	string `json:"password"`
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

	usr := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	respondWithJSON(rw, 200, usr)
}

func (cfg *apiConfig) handlerCreateChirp(rw http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	params := parameters{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(rw, 400, "Chirp is too long")
		return
	} else {
		params.Body = replaceProfaneWords(params.Body, getProfaneWords())
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams(params))
	if err != nil {
		log.Printf("Error creating the chirp on the database: %s", err)
		respondWithError(rw, 500, "Can't create chirp")
		return
	}

	mappedChirp := Chirp(chirp)
	respondWithJSON(rw, 201, mappedChirp)
	return
}

func (cfg *apiConfig) handlerGetChirps(rw http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.ListChirps(req.Context())
	if err != nil {
		log.Printf("Error retrieving the chirps from the database: %s", err)
		respondWithError(rw, 500, "Can't retrieve chirps")
		return
	}

	mappedChirps := []Chirp{}
	for _, c := range chirps {
		mappedChirps = append(mappedChirps, Chirp(c))
	}
	respondWithJSON(rw, 200, mappedChirps)
	return
}

func (cfg *apiConfig) handlerGetChirp(rw http.ResponseWriter, req *http.Request) {
	cid := req.PathValue("chirpID")
	if cid == "" {
		respondWithError(rw, 400, "Missing chirp ID in request")
		return
	}

	parsedUUID, err := uuid.Parse(cid)
	if err != nil {
		log.Printf("The ID of the request can't be parsed into a UUID")
		respondWithError(rw, 400, "Invalid ID")
		return
	}

	chirp, err := cfg.db.GetChirp(req.Context(), parsedUUID)
	if err != nil {
		respondWithError(rw, 404, "Chirp not found")
		return
	}

	mappedChirp := Chirp(chirp)
	respondWithJSON(rw, 200, mappedChirp)
	return
}

func (cfg *apiConfig) handlerGetHits(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	body := fmt.Sprintf(
		`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, cfg.fileserverHits.Load())
	rw.Write([]byte(body))
}

func (cfg *apiConfig) handlerReset(rw http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		rw.WriteHeader(403)
		return
	}
	cfg.fileserverHits.Store(0)
	cfg.db.DeleteUsers(req.Context())
	rw.WriteHeader(http.StatusOK)
	return
}

func handlerValidateChirp(rw http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	params := parameters{}

	type returnValues struct {
		Error string `json:"error"`
		Cleaned_Body string `json:"cleaned_body"`
	}
	resBody := returnValues{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(rw, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(rw, 400, "Chirp is too long")
	} else {
		resBody.Cleaned_Body = replaceProfaneWords(params.Body, getProfaneWords())
		respondWithJSON(rw, 200, resBody)
	}

	return
}

func respondWithError(rw http.ResponseWriter, code int, msg string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	data, _ := json.Marshal(struct{Error string `json:"error"`}{Error: msg})
	rw.Write(data)
	return
}

func respondWithJSON(rw http.ResponseWriter, code int, payload interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling payload: %s", err)
		rw.WriteHeader(500)
		return
	}
	rw.WriteHeader(code)
	rw.Write(data)
	return
}

func getProfaneWords() []string {
	return []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
}

func replaceProfaneWords(sentence string, profaneWords []string) string {
	words := strings.Split(sentence, " ")
	cleanWords := []string{}
	for _, w := range words {
		profane := false
		for _, p := range profaneWords {
			if strings.ToLower(w) == p {
				profane = true
				break
			}
		}
		if profane {
			cleanWords = append(cleanWords, "****")
		} else {
			cleanWords = append(cleanWords, w)
		}
	}
	cleanSentence := strings.Join(cleanWords, " ")
	return cleanSentence
}