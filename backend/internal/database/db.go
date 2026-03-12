// db.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

type Website struct {
	ID          int       `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
}

var (
	ErrWebsiteAlreadyExists      = errors.New("website with this URL already exists")
	ErrSubscriptionAlreadyExists = errors.New("already subscribed to this website")
)

func ConnectToDb() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("[!] WARNING: .env file not found, falling back to system environment variables.")
	}

	host := "pg_db"
	port := 5432
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("[-] FATAL: Failed to open database connection: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("[-] FATAL: Database is unreachable (ping failed): ", err)
	}

	log.Println("[+] Successfully connected to the PostgreSQL database!")
}

func AddWebsite(ownerID, name, url, description string, isPublic bool) error {
	query := `
		INSERT INTO websites (owner_id, name, url, description, is_public) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (owner_id, url) DO NOTHING`

	result, err := DB.Exec(query, ownerID, name, url, description, isPublic)
	if err != nil {
		log.Printf("[-] ERROR: Failed to insert website %s: %v\n", url, err)
		return err
	}

	// ON CONFLICT DO NOTHING silently skips the insert — check if a row was actually written
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Printf("[!] DUPLICATE: Website %s already exists for this owner.\n", url)
		return ErrWebsiteAlreadyExists
	}

	log.Printf("[+] SUCCESS: Website %s added to monitoring.\n", name)
	return nil
}

func SubscribeToWebsite(userID string, websiteID int) error {
	query := `
		INSERT INTO subscriptions (user_id, website_id) 
		VALUES ($1, $2)
		ON CONFLICT (user_id, website_id) DO NOTHING`

	result, err := DB.Exec(query, userID, websiteID)
	if err != nil {
		log.Printf("[-] ERROR: Failed to subscribe user %s to website ID %d: %v\n", userID, websiteID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Printf("[!] DUPLICATE: User %s is already subscribed to website ID %d.\n", userID, websiteID)
		return ErrSubscriptionAlreadyExists
	}

	return nil
}