package main

import (
	"log"
)

func main() {
	// 1. Start the Database (The foundation)
	InitDB()

	// 2. Set up the Network (The interface)
	r := SetupRouter()

	// 3. Launch!
	log.Println("[+] Watcher server is roaring on port 8080...")
	
	// Default port is 8080. If you want another, use r.Run(":3000")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("[-] Server crashed:", err)
	}
}