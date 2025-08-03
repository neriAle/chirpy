package main

import(
	"net/http"
)

func startServer() {
	const filepathRoot = "."
	const port = "9090"

	servemux := http.NewServeMux()

	servemux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	servemux.HandleFunc("/healthz/", handlerHealthz)

	server := http.Server{}
	server.Addr = ":" + port
	server.Handler = servemux

	server.ListenAndServe()
}

func handlerHealthz(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(http.StatusText(http.StatusOK)))
}