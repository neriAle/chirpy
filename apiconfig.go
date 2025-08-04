package main

import(
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		Valid bool `json:"valid"`
	}
	resBody := returnValues{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %w", err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(500)

		resBody.Error = "Something went wrong"
		data, err1 := json.Marshal(resBody)
		if err1 == nil {
			rw.Write(data)
		}

		return
	}

	if len(params.Body) > 140 {
		resBody.Error = "Chirp is too long"
		rw.WriteHeader(400)
	} else {
		resBody.Valid = true
		rw.WriteHeader(200)
	}
	data, err := json.Marshal(resBody)
	if err == nil {
		rw.Write(data)
	}

	return
}