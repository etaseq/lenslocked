package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

//--------------------------------------------------------------------------
// bash
// # Hash a password
// go build cmd/bcrypt/bcrypt.go

// # Produces binary named "bcrypt"
// ./bcrypt hash "some password here"

// # Compare a hash and password
// go run cmd/bcrypt/bcrypt.go compare "some password here" "some hash here"
//---------------------------------------------------------------------------

// GenerateFromPassword needs a byte slice for password so I need to type cast it.
// The cost parameter is the extra time it will take to generate the hash so that
// brute force attacks will be prevented. It determines the number of iterations
// used to hash the password.
// Higher cost = more computational time = harder for attackers to crack.
// Just like the salt (which is for preventing Rainbow table attacks and is
// automatically produced by bcrypt), it is embedded inside the produced hash.
// Here I am using DefaultCost but I can also use custom values.
// I can change the cost value later in time and since it is embedded in the hash,
// the system would have no problem identifying older passwords which have the
// previous cost value.
func hash(password string) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing: %v\n", password)
		return
	}
	// bcrypt uses base-64 encoding so I can type cast the byte slice back to a string
	hash := string(hashedBytes)
	fmt.Println(hash)
}

func compare(password, hash string) {
	fmt.Printf("Password: %q\nHash: %q\n", password, hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password is invalid: %v\n", password)
		return
	}
	fmt.Println("Password is correct!")
}

func main() {
	switch os.Args[1] {
	case "hash":
		hash(os.Args[2])
	case "compare":
		compare(os.Args[2], os.Args[3])
	default:
		fmt.Printf("Invalid command: %v\n", os.Args[1])
	}
}
