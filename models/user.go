package models

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
}

type UserService struct {
	DB *sql.DB
}

// Create a new user. Notice that I return a *User pointer.
// In the case of an error, returning a *User allows you to return 
// nil (which represents "no valid User object") instead of an 
// empty User object. This gives you a clear signal that the user 
// creation failed.
func (us *UserService) Create(email, password string) (*User, error) {
	// Postgres is case sensitive so convert all email letters
	// to lower case to prevent duplicate entries.
	email = strings.ToLower(email)

	// Follow the same process as in cmd/bcrypt/bcrypt.go to
	// generate a hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	passwordHash := string(hashedBytes)

	user := User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	// Insert the new user to the database and return a *sql.Row.
	row := us.DB.QueryRow(`
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2) RETURNING id`, email, passwordHash)

	// Extract the new user's id and store it to the user.ID field
	err = row.Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, err
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)

	user := User{
		Email: email,
	}

	row := us.DB.QueryRow(`
		SELECT id, password_hash
		FROM users WHERE email=$1`, email)

	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return &user, nil
}
