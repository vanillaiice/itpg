package itpg

import (
	"encoding/json"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const SessionTokenValidity = time.Second * 120

func Signup(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	err = isEmptyStr(w, creds.Username, creds.Password)
	if err != nil {
		return
	}

	exists, err := userExists(w, creds.Username)
	if err != nil {
		return
	}
	if exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = addUser(w, creds.Username, creds.Password)
	if err != nil {
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	match, err := checkPassword(w, creds.Username, creds.Password)
	if err != nil {
		return
	}
	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// check if session token expired ?

	exists, err := sessionExistsByUsername(w, creds.Username)
	if err != nil {
		return
	}
	if exists {
		err = deleteSessionTokenByUsername(w, creds.Username)
	}

	expiry := time.Now().Add(SessionTokenValidity)
	sessionToken, err := addSession(w, creds.Username, expiry)
	if err != nil {
		return
	}

	setCookie(w, sessionToken, expiry)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c, err := getCookie(w, r)
	if err != nil {
		return
	}

	exists, err := sessionExists(w, c.Value)
	if err != nil {
		return
	}
	if !exists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = deleteSessionToken(w, c.Value)
	if err != nil {
		return
	}

	setCookie(w, "", time.Now())
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := getCookie(w, r)
	if err != nil {
		return
	}

	expired, err := checkSessionTokenExpiry(w, c.Value)
	if err != nil {
		return
	}
	if expired {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expiry := time.Now().Add(SessionTokenValidity)
	newSessionToken, err := refreshSessionToken(w, c.Value, expiry)
	if err != nil {
		return
	}

	setCookie(w, newSessionToken, expiry)
}

func Greet(w http.ResponseWriter, r *http.Request) {
	c, err := getCookie(w, r)
	if err != nil {
		return
	}
	expired, err := checkSessionTokenExpiry(w, c.Value)
	if err != nil {
		return
	}
	if expired {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(&Message{Message: "Sup mate ?"})
	/*
			c, err := r.Cookie("session_token")
			if err != nil {
				if err == http.ErrNoCookie {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
				return
			}
			sessionToken := c.Value

			exists, err := authDB.SessionExists(sessionToken)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
				return
			}
			if !exists {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

		expired, err := authDB.CheckSessionTokenExpiry(sessionToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&Response{Status: Err, Message: err.Error()})
			return
		}
		if expired {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&Response{Status: Err, Message: "session token expired"})
			return
		}
	*/
}
