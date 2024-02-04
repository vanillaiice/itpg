package itpg

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthDB is a struct representing the authentication database and its methods.
type AuthDB struct {
	db *sql.DB
}

// Close closes the authentication database connection.
func (db *AuthDB) Close() error {
	return db.db.Close()
}

// NewAuthDB initializes a new authentication database connection and sets up the necessary tables if they don't exist.
func NewAuthDB(dbPath string) (*AuthDB, error) {
	db := &AuthDB{}
	sqldb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.db = sqldb

	stmt := []string{
		"PRAGMA foreign_keys = ON",
		"CREATE TABLE IF NOT EXISTS Users(username TEXT PRIMARY KEY NOT NULL CHECK(username != ''), passwordhash TEXT(72) NOT NULL)",
		"CREATE TABLE IF NOT EXISTS Sessions(username TEXT PRIMARY KEY NOT NULL, sessiontoken TEXT NOT NULL, sessiontokenexpiry INTEGER NOT NULL, FOREIGN KEY(username) REFERENCES Users(username))",
	}
	for _, s := range stmt {
		_, err := execStmt(db.db, s)
		if err != nil {
			return nil, err
		}
	}
	return db, err
}

// AddUser adds a new user to the authentication database with the provided username and password.
func (db *AuthDB) AddUser(username, password string) (err error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	stmt := fmt.Sprintf("INSERT INTO Users(username, passwordhash) VALUES(%q, %q)", username, string(passwordHash))
	_, err = execStmt(db.db, stmt)
	return
}

// AddSession adds a new session to the authentication database with the provided username, session token, and session token expiry time.
func (db *AuthDB) AddSession(username, sessionToken string, sessinTokenExpiry time.Time) (err error) {
	stmt := fmt.Sprintf("INSERT INTO Sessions(username, sessiontoken, sessiontokenexpiry) VALUES(%q, %q, %d)", username, sessionToken, sessinTokenExpiry.Unix())
	_, err = execStmt(db.db, stmt)
	return
}

// UserExists checks if a user with the given username exists in the authentication database.
func (db *AuthDB) UserExists(username string) (exists bool, err error) {
	var u string
	stmt := fmt.Sprintf("SELECT username FROM Users WHERE username = %q", username)
	err = db.db.QueryRow(stmt).Scan(&u)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SessionExists checks if a session with the given session token exists in the authentication database.
func (db *AuthDB) SessionExists(sessionToken string) (exists bool, err error) {
	var s string
	stmt := fmt.Sprintf("SELECT sessiontoken FROM Sessions WHERE sessiontoken = %q", sessionToken)
	err = db.db.QueryRow(stmt).Scan(&s)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SessionExists checks if a session with the specified username exists in the authentication database.
func (db *AuthDB) SessionExistsByUsername(username string) (exists bool, err error) {
	var u string
	stmt := fmt.Sprintf("SELECT sessiontoken FROM Sessions WHERE username = %q", username)
	err = db.db.QueryRow(stmt).Scan(&u)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CheckPassword checks if the provided password matches the hashed password stored in the authentication database for the given username.
func (db *AuthDB) CheckPassword(username, password string) (mactch bool, err error) {
	var hashedPassword string
	stmt := fmt.Sprintf("SELECT passwordhash FROM Users WHERE username = %q", username)
	err = db.db.QueryRow(stmt).Scan(&hashedPassword)
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// CheckSessionTokenExpiry checks if the session token has expired in the authentication database.
func (db *AuthDB) CheckSessionTokenExpiry(sessionToken string) (expired bool, err error) {
	var sessionTokenExpiry int64
	stmt := fmt.Sprintf("SELECT sessiontokenexpiry FROM Sessions WHERE sessiontoken = %q", sessionToken)
	err = db.db.QueryRow(stmt).Scan(&sessionTokenExpiry)
	if err != nil {
		return false, err
	}
	timeSessionTokenExpiry := time.Unix(sessionTokenExpiry, 0)
	expired = timeSessionTokenExpiry.Before(time.Now())
	return
}

// RefreshSessionToken updates the session token and its expiry time in the authentication database.
func (db *AuthDB) RefreshSessionToken(sessionToken, newSessionToken string, newSessionTokenExpiry time.Time) (n int64, err error) {
	stmt := fmt.Sprintf("UPDATE Sessions SET sessiontoken = %q, sessiontokenexpiry = %d WHERE sessiontoken = %q", newSessionToken, newSessionTokenExpiry.Unix(), sessionToken)
	n, err = execStmt(db.db, stmt)
	return
}

// DeleteSessionToken deletes a session token from the authentication database.
func (db *AuthDB) DeleteSessionToken(sessionToken string) (n int64, err error) {
	stmt := fmt.Sprintf("DELETE FROM Sessions WHERE sessiontoken = %q", sessionToken)
	n, err = execStmt(db.db, stmt)
	return
}

// DeleteSessionToken deletes a session token with the specified username from the authentication database.
func (db *AuthDB) DeleteSessionTokenByUsername(username string) (n int64, err error) {
	stmt := fmt.Sprintf("DELETE FROM Sessions WHERE username = %q", username)
	n, err = execStmt(db.db, stmt)
	return
}
