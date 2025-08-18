package main

import(
	"database/sql"
	"log"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/neriAle/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits 	atomic.Int32
	db 				*database.Queries
	platform 		string
	tokenSecret 	string
	polka_key		string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	tokenSecret := os.Getenv("TOKEN_SECRET")
	if tokenSecret == "" {
		log.Fatal("TOKEN_SECRET must be set")
	}
	polka := os.Getenv("POLKA_KEY")
	if polka == "" {
		log.Fatal("POLKA_KEY must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
		platform: platform,
		tokenSecret: tokenSecret,
		polka_key: polka,
	}

	startServer(apiCfg)
}