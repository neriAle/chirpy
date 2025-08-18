package main

import(
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/neriAle/chirpy/internal/auth"
	"github.com/neriAle/chirpy/internal/database"
)

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

	params.UserID = userId

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
	s := req.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error
	if s == "" {
		chirps, err = cfg.db.ListChirps(req.Context())
	} else {
		parsedUUID, err := uuid.Parse(s)
		if err != nil {
			log.Printf("The ID of the user can't be parsed into a UUID")
			respondWithError(rw, 400, "Invalid user ID")
			return
		}
		chirps, err = cfg.db.GetChirpsByAuthor(req.Context(), parsedUUID)
	}

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
}

func (cfg *apiConfig) handlerDeleteChirp(rw http.ResponseWriter, req *http.Request) {
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

	if chirp.UserID != userId {
		respondWithError(rw, 403, "Not authorized to delete this chirp")
		return
	}

	err = cfg.db.DeleteChirp(req.Context(), parsedUUID)
	if err != nil {
		respondWithError(rw, 500, "Error deleting the chirp")
		return
	}

	rw.WriteHeader(204)
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