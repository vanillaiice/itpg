package itpg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func checkCookieExpiry(w http.ResponseWriter, r *http.Request) (err error) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cookieExpiry, err := userState.Users().Get(username, "cookie-expiry")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cookieExpiryTime, err := time.Parse(time.UnixDate, cookieExpiry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if time.Now().After(cookieExpiryTime) {
		if userState.IsLoggedIn(username) {
			userState.Logout(username)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return fmt.Errorf("cookie expired")
	}
	return
}

func checkCookieExpiryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expired := checkCookieExpiry(w, r); expired != nil {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkUserAlreadyGradedMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, err := userState.UsernameCookie(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		courseCode := r.FormValue("code")
		_, err = userState.Users().Get(username, courseCode)
		if err == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
		defer func() {
			userState.Users().Set(username, courseCode, "")
		}()
	})
}
