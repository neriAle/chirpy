package main

import(
	"github.com/neriAle/chirpy/internal/database"
	"net/http"
	"sync/atomic"
)

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

	server := &http.Server{
		Addr:    ":" + port,
		Handler: servemux,
	}

	server.ListenAndServe()
}