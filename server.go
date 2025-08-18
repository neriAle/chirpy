package main

import(
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/neriAle/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits 	atomic.Int32
	db 				*database.Queries
	platform 		string
	tokenSecret 	string
}

type User struct {
	ID        	uuid.UUID 	`json:"id"`
	CreatedAt 	time.Time 	`json:"created_at"`
	UpdatedAt 	time.Time 	`json:"updated_at"`
	Email     	string    	`json:"email"`
	IsChirpyRed	bool		`json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string	`json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func startServer(dbq *database.Queries, pf string, ts string) {
	const filepathRoot = "."
	const port = "9090"

	servemux := http.NewServeMux()
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbq,
		platform: pf,
		tokenSecret: ts,
	}

	servemux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	servemux.HandleFunc("GET /api/healthz", handlerHealthz)
	servemux.HandleFunc("GET /admin/metrics", apiCfg.handlerGetHits)
	servemux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	servemux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	servemux.HandleFunc("POST /api/login", apiCfg.handlerLoginUser)
	servemux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	servemux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	servemux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	servemux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshJWT)
	servemux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefreshToken)
	servemux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	servemux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	servemux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUpgradeUser)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: servemux,
	}

	server.ListenAndServe()
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