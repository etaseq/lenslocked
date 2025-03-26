package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/etaseq/lenslocked/rand"
)

const (
	// The minimum number of bytes to be used for each session token.
	MinBytesPerToken = 32
)

type Session struct {
	ID     int
	UserID int
	// Token is only set when creating a new session.
	// When look up a session this will be left empty,
	// as I only store the hash of a session token in our
	// database and we cannot reverse it into a raw token
	Token     string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
	// Bytes per token is used to determine how many bytes to use when generating
	// each session token. If this value is not set or is less than the
	// MinBytesPerToken const it will be ignored and MinBytesPerToken will be used.
	BytesPerToken int
}

func (ss *SessionService) Create(userID int) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}

	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: ss.hash(token),
	}

	// The Create method is also used for Signin so I will need
	// to be able to update the sessions too.
	// 1. Try to UPDATE the user's session
	// 2. If err, create new session
	//row := ss.DB.QueryRow(`
	//	UPDATE sessions
	//	SET token_hash = $2
	//	WHERE user_id = $1
	//	RETURNING id;`, session.UserID, session.TokenHash)
	//err = row.Scan(&session.ID)

	//if err == sql.ErrNoRows {
	//	row = ss.DB.QueryRow(`
	//		INSERT INTO sessions (user_id, token_hash)
	//		VALUES ($1, $2)
	//		RETURNING id;`, session.UserID, session.TokenHash)
	//	err = row.Scan(&session.ID)
	//}

	// Short version of the 1. and 2. using Postgres ON CONFLICT
	row := ss.DB.QueryRow(`
		INSERT INTO sessions (user_id, token_hash)
		VALUES ($1, $2) ON CONFLICT (user_id) DO
		UPDATE
		SET token_hash = $2
		RETURNING id;`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)

	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &session, nil
}

func (ss *SessionService) User(token string) (*User, error) {
	// 1. Hash the session token
	tokenHash := ss.hash(token)

	// 2. Query for the session with that hash
	//var user User
	//row := ss.DB.QueryRow(`
	//	SELECT user_id
	//	FROM sessions
	//	WHERE token_hash = $1`, tokenHash)
	//err := row.Scan(&user.ID)
	//if err != nil {
	//	return nil, fmt.Errorf("user: %w", err)
	//}

	//// 3. Using the UserID from the session, query for that User
	//row = ss.DB.QueryRow(`
	//	SELECT email, password_hash
	//	FROM users WHERE id = $1;`, user.ID)
	//err = row.Scan(&user.Email, &user.PasswordHash)
	//if err != nil {
	//	return nil, fmt.Errorf("user: %w", err)
	//}

	// Short version of 2. and 3. using JOIN
	var user User
	row := ss.DB.QueryRow(`
		SELECT users.id,
			users.email,
			users.password_hash
		FROM sessions
			JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1;`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	// 4. Return the user
	return &user, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)

	// I don't need something to be returned from the query
	// so I can use Exec.
	_, err := ss.DB.Exec(`
		DELETE FROM sessions
		WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// I do hash instead of Hash because I do not want this function
// to be used outside of this scope.
func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	// Notice that Sum256 returns an array so I need to use [:]
	// to convert the tokenHash array to a slice to feed it to
	// EncodeToString
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
