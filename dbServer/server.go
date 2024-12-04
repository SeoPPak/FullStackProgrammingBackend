package main

import (
	jwt "dbserver/auth"
	"log"
)

func main() {
	if err := jwt.InitializeKeys(); err != nil {
		log.Fatalf("Failed to initialize RSA keys: %v", err)
	}

	r := Setup()

	r.Run(":5000")
}