package main

import(
	"net/http"
	"sync/atomic"
)

func startServer() {
	const filepathRoot = "."
	const port = "9090"

	servemux := http.NewServeMux()
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	servemux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	servemux.HandleFunc("GET /api/healthz", handlerHealthz)
	servemux.HandleFunc("GET /admin/metrics", apiCfg.handlerGetHits)
	servemux.HandleFunc("POST /admin/reset", apiCfg.handlerResetHits)
	servemux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: servemux,
	}

	server.ListenAndServe()
}