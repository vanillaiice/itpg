package itpg

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/xyproto/permissionbolt/v2"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	err = permissionbolt.ValidUsernamePassword(creds.Username, creds.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hasUser := userState.HasUser(creds.Username)
	if hasUser {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userState.AddUser(creds.Username, creds.Password, creds.Email)
}

func Login(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	hasUser := userState.HasUser(creds.Username)
	if !hasUser {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	correct := userState.CorrectPassword(creds.Username, creds.Password)
	if !correct {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	/*
		isLoggedIn := userState.IsLoggedIn(creds.Username)
		if isLoggedIn {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	*/

	userState.Users().Set(
		creds.Username,
		"cookie-expiry",
		time.Now().Add(cookieTimeout).Format(time.UnixDate),
	)

	userState.Login(w, creds.Username)
}

func Admin(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}
	correct := userState.CorrectPassword(creds.Username, creds.Password)
	if !correct {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userState.AddUser(creds.Username, creds.Password, creds.Email)
	userState.Login(w, creds.Username)
	userState.SetAdminStatus(creds.Username)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isLoggedIn := userState.IsLoggedIn(username)
	if !isLoggedIn {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	userState.Logout(username)
}

func ClearCookie(w http.ResponseWriter, r *http.Request) {
	userState.ClearCookie(w)
}

func RefreshCookie(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userState.Users().Set(
		username,
		"cookie-expiry",
		time.Now().Add(cookieTimeout).Format(time.UnixDate),
	)

	userState.Login(w, username)
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userState.Users().DelKey(username, "cookie-expiry")
	userState.RemoveUser(username)
}

func Greet(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(&Message{Message: "Sup mate ?"})
}
