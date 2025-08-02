package main

import(
	"net/http"
)

func main() {
	servemux := http.NewServeMux()
	servemux.Handle("/", http.FileServer(http.Dir(".")))

	server := http.Server{}
	server.Addr = ":9090"
	server.Handler = servemux

	server.ListenAndServe()
}