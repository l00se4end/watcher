package main

import (
	"log"
	
	"watcher/internal/database"
	"watcher/internal/server"
)

func main() {
	log.Println("[*] Server Starting...")

	database.ConnectToDb()

	r := server.SetupRouter()

	log.Println("[+] Web server running on port 8080")
	
	if err := r.Run(":8080"); err != nil {
		log.Fatal("[-] FATAL: ", err)
	}
}