package main

import (
	"fmt"
	"os"
	
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hashtoken <token>")
		os.Exit(1)
	}
	
	token := os.Args[1]
	
	// Hash with cost 10 (same as registration system)
	hash, err := bcrypt.GenerateFromPassword([]byte(token), 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Token: %s\n", token)
	fmt.Printf("Hash:  %s\n", string(hash))
	
	// Verify it works
	err = bcrypt.CompareHashAndPassword(hash, []byte(token))
	if err != nil {
		fmt.Println("Verification: FAILED")
	} else {
		fmt.Println("Verification: OK")
	}
}
