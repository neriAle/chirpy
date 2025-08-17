package main

import(
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/neriAle/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	startServer(dbQueries, platform, tokenSecret)
}