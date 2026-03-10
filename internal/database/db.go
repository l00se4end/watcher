package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DB represents the global connection pool.
// Exported so other packages can use the initialized connection.
var DB *sql.DB

// Website represents the data model for monitored targets.
type Website struct {
	ID          int
	OwnerID     string // UUID mapped from Authentik JWT/Session
	Name        string
	URL         string
	IP          string
	Description string
	IsPublic    bool
}

// ConnectToDb initializes the PostgreSQL connection pool.
func ConnectToDb() {
	// Load .env file from the project root.
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("[!] WARNING: .env file not found, falling back to system environment variables.")
	}

	host := "pg_db" // Docker compose service name
	port := 5432
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Initialize the global DB connection pool.
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("[-] FATAL: Failed to open database connection: ", err)
	}

	// Verify the connection is established.
	if err = DB.Ping(); err != nil {
		log.Fatal("[-] FATAL: Database is unreachable (ping failed): ", err)
	}

	log.Println("[+] Successfully connected to the PostgreSQL database! 🚀")
}

// AddWebsite inserts a new target into the database.
// Uses parameterized queries ($1, $2...) to prevent SQL injection.
func AddWebsite(ownerID, name, url, ip, description string, isPublic bool) error {
	query := `
		INSERT INTO websites (owner_id, name, url, ip, description, is_public) 
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := DB.Exec(query, ownerID, name, url, ip, description, isPublic)
	if err != nil {
		log.Printf("[-] ERROR: Failed to insert website %s: %v\n", url, err)
		return err
	}
	
	log.Printf("[+] SUCCESS: Website %s added to monitoring.\n", name)
	return nil
}

// SubscribeToWebsite links a user to a specific website for uptime notifications.
func SubscribeToWebsite(userID string, websiteID int) error {
	query := `
		INSERT INTO subscriptions (user_id, website_id) 
		VALUES ($1, $2)`

	_, err := DB.Exec(query, userID, websiteID)
	if err != nil {
		log.Printf("[-] ERROR: Failed to subscribe user %s to website ID %d: %v\n", userID, websiteID, err)
		return err
	}

	return nil
}