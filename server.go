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
	servemux.HandleFunc("/healthz/", handlerHealthz)
	servemux.HandleFunc("/metrics/", apiCfg.handlerGetHits)
	servemux.HandleFunc("/reset/", apiCfg.handlerResetHits)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: servemux,
	}

	server.ListenAndServe()
}