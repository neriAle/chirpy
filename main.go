package main

import(
	"github.com/neriAle/chirpy/internal/database"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	startServer(dbQueries)
}