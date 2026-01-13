package main

import (
	"log"
	"os"

	_ "modernc.org/sqlite"

	"goodreads-clone.com/internal/business"
	"goodreads-clone.com/internal/server"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "goodreads.db"
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	b, err := business.New(business.Config{
		DatabasePath: dbPath,
	})
	if err != nil {
		log.Fatalf("Failed to initialize business layer: %v", err)
	}
	defer b.Close()

	srv, err := server.New(b)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	log.Printf("Server starting on %s", addr)
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
