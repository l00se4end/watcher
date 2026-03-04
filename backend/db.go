package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" 
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./uptime.db")
	if err != nil {
		log.Fatal("[-] Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("[-] Database is unreachable:", err)
	}

	// Dodao sam 'status' kolonu jer će nam trebati za onaj tvoj 'UP/DOWN' vizual
	query := `
	CREATE TABLE IF NOT EXISTS Monitors(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		status TEXT DEFAULT 'UP'
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("[-] Failed to create table:", err)
	}
	log.Println("[+] Database initialized and ready 🚀")
}

// SaveMonitor sada vraća (ID, error) - ključno za HTMX dinamiku!
func SaveMonitor(name, url string) (int64, error) {
	// Koristimo direktno Exec jer nam treba Result objekt za LastInsertId
	res, err := db.Exec("INSERT INTO Monitors (name, url) VALUES (?, ?)", name, url)
	if err != nil {
		return 0, err
	}

	// Uzimamo ID koji je SQLite upravo generirao
	return res.LastInsertId()
}

// Nova funkcija za brisanje - spaja se na hx-delete
func DeleteMonitor(id string) error {
	_, err := db.Exec("DELETE FROM Monitors WHERE id = ?", id)
	return err
}

func GetAllMonitors() ([]Monitor, error) {
	rows, err := db.Query("SELECT id, name, url FROM Monitors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []Monitor
	for rows.Next() {
		var m Monitor
		if err := rows.Scan(&m.ID, &m.Name, &m.URL); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}