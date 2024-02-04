package itpg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

func decodeCredentials(w http.ResponseWriter, r *http.Request) (*Credentials, error) {
	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return &credentials, err
}

func isEmptyStr(w http.ResponseWriter, str ...string) error {
	for _, s := range str {
		if s == "" {
			w.WriteHeader(http.StatusBadRequest)
			return fmt.Errorf("got empty str")
		}
	}
	return nil
}

func checkPassword(w http.ResponseWriter, username, password string) (bool, error) {
	match, err := authDB.CheckPassword(username, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}
	return match, nil
}

func checkSessionTokenExpiry(w http.ResponseWriter, sessionToken string) (bool, error) {
	expired, err := authDB.CheckSessionTokenExpiry(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}
	return expired, nil
}

func userExists(w http.ResponseWriter, username string) (bool, error) {
	exists, err := authDB.UserExists(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}
	return exists, nil
}

func sessionExists(w http.ResponseWriter, sessionToken string) (bool, error) {
	exists, err := authDB.SessionExists(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}
	return exists, nil
}

func sessionExistsByUsername(w http.ResponseWriter, username string) (bool, error) {
	exists, err := authDB.SessionExistsByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}
	return exists, nil
}

func addUser(w http.ResponseWriter, username, password string) error {
	err := authDB.AddUser(username, password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func addSession(w http.ResponseWriter, username string, expiry time.Time) (string, error) {
	sessionToken, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", err
	}
	err = authDB.AddSession(username, sessionToken.String(), expiry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", err
	}
	return sessionToken.String(), nil
}

func refreshSessionToken(w http.ResponseWriter, sessionToken string, expiry time.Time) (string, error) {
	newSessionToken, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", err
	}
	n, err := authDB.RefreshSessionToken(sessionToken, newSessionToken.String(), expiry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", err
	}
	if n == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return "", fmt.Errorf("no rows affected")
	}
	return newSessionToken.String(), nil
}

func deleteSessionToken(w http.ResponseWriter, sessionToken string) error {
	n, err := authDB.DeleteSessionToken(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if n == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func deleteSessionTokenByUsername(w http.ResponseWriter, username string) error {
	n, err := authDB.DeleteSessionTokenByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if n == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func getCookie(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return c, nil
}

func setCookie(w http.ResponseWriter, sessionToken string, expiry time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiry,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
