package main

import(
	"fmt"
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
	rw.Header().Add("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)
	body := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	rw.Write([]byte(body))
}

func (cfg *apiConfig) handlerResetHits(rw http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	rw.WriteHeader(http.StatusOK)
}