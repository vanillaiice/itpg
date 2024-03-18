package itpg

import (
	"context"
	"itpg/responses"
	"net/http"
	"time"
)

// UsernameContextKey is the key in the request's context to set
// the username for use in subsequent middleware.
const UsernameContextKey = "username"

// CookieExpiryUserStateKey is the key in the Userstate database
// use to retrieve the expiry time of a session cookie.
const CookieExpiryUserStateKey = "cookie-expiry"

// checkCookieExpiry checks if the user's session cookie has expired.
// If the cookie has expired, it logs out the user, writes an Unauthorized response, and returns an error.
// It returns nil if the cookie is valid and has not expired.
func checkCookieExpiry(username string) error {
	cookieExpiry, err := UserState.Users().Get(username, CookieExpiryUserStateKey)
	if err != nil {
		return responses.ErrInvalidCookie
	}

	cookieExpiryTime, err := time.Parse(time.UnixDate, cookieExpiry)
	if err != nil {
		return responses.ErrInternal
	}

	if time.Now().After(cookieExpiryTime) {
		if UserState.IsLoggedIn(username) {
			UserState.Logout(username)
		}
		return responses.ErrExpiredCookie
	}

	return nil
}

// DummyMiddleware is middleware that does nothing.
// It is used to wrap the go-chi/httprate limiter around a handler
func DummyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}

// checkCookieExpiryMiddleware is a middleware that checks if the user's session cookie has expired.
// If the cookie has expired, it writes an Unauthorized response and returns.
// It calls the next handler if the cookie is valid and has not expired.
func checkCookieExpiryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := UserState.UsernameCookie(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			responses.ErrInvalidCookie.WriteJSON(w)
			return
		}

		if expired := checkCookieExpiry(username); expired == nil {
			r = r.WithContext(context.WithValue(r.Context(), UsernameContextKey, username))
			next.ServeHTTP(w, r)
		} else {
			switch err {
			case responses.ErrInvalidCookie:
				w.WriteHeader(http.StatusUnauthorized)
				responses.ErrInvalidCookie.WriteJSON(w)
			case responses.ErrInternal:
				w.WriteHeader(http.StatusInternalServerError)
				responses.ErrInternal.WriteJSON(w)
			case responses.ErrExpiredCookie:
				w.WriteHeader(http.StatusForbidden)
				responses.ErrExpiredCookie.WriteJSON(w)
			}
		}
	}
}

// checkConfirmedMiddleware is a middleware that checks if the user is confirmed.
// If the user is not confirmed, it writes a Forbidden response.
// It calls the next handler if the user is confirmed.
func checkConfirmedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, ok := r.Context().Value(UsernameContextKey).(string)
		if !ok || username == "" {
			w.WriteHeader(http.StatusInternalServerError)
			responses.ErrInternal.WriteJSON(w)
			return
		}

		if !UserState.IsConfirmed(username) {
			w.WriteHeader(http.StatusUnauthorized)
			responses.ErrNotConfirmed.WriteJSON(w)
			return
		}

		next.ServeHTTP(w, r)
	}
}
