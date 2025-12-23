package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run test_password.go <plain-password> <bcrypt-hash>")
		fmt.Println("\nExample:")
		fmt.Println("  go run test_password.go mypassword '$2a$10$xxx...'")
		fmt.Println("\nOr to just generate a new hash:")
		fmt.Println("  go run hash_password.go <plain-password>")
		os.Exit(1)
	}

	plainPassword := os.Args[1]
	hash := os.Args[2]

	// Test if the plain password matches the hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
	if err == nil {
		fmt.Printf("✅ SUCCESS! Password '%s' matches the hash.\n", plainPassword)
		fmt.Println("\nYou can login with:")
		fmt.Printf("  Username: %s\n", os.Getenv("ADMIN_USERNAME"))
		fmt.Printf("  Password: %s\n", plainPassword)
	} else {
		fmt.Printf("❌ FAILED! Password '%s' does NOT match the hash.\n", plainPassword)
		fmt.Println("\nError:", err)
		fmt.Println("\nTo generate a new hash for this password, run:")
		fmt.Printf("  go run hash_password.go %s\n", plainPassword)
	}
}
