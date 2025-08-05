package main

import(
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(rw, req)
	})
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

func (cfg *apiConfig) handlerResetHits(rw http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	rw.WriteHeader(http.StatusOK)
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
		log.Printf("Error decoding parameters: %w", err)
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
		log.Printf("Error marshalling payload: %w", err)
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