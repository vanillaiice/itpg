package itpg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

// isEmptyStr checks if any of the provided strings are empty.
func isEmptyStr(w http.ResponseWriter, str ...string) (err error) {
	for _, s := range str {
		if s == "" {
			w.WriteHeader(http.StatusBadRequest)
			ErrEmptyValue.WriteJSON(w)
			return ErrEmptyValue.Error()
		}
	}
	return
}

// decodeCredentials decodes JSON data from the request body into a Credentials struct.
func decodeCredentials(w http.ResponseWriter, r *http.Request) (*Credentials, error) {
	var credentials Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		ErrBadRequest.WriteJSON(w)
		return nil, err
	}
	return &credentials, nil
}

// extractDomain extracts the domain part from an email address.
// It takes an email address string as input and returns the domain part.
// If the email address is in an invalid format (e.g., missing "@" symbol),
// it returns an empty string.
func extractDomain(email string) (domain string, err error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return domain, fmt.Errorf("invalid email format")
	}
	domain = parts[1]
	return domain, err
}

// validAllowedDomains checks if the list of allowed mail domains is empty.
func validAllowedDomains(domains []string) (err error) {
	if len(domains) == 0 {
		return fmt.Errorf("got empty list of allowed mail domains")
	}
	return
}

// checkDomainAllowed checks if the given domain is allowed based on the list of allowed mail domains.
func checkDomainAllowed(domain string) (err error) {
	if AllowedMailDomains[0] == "*" {
		return
	}
	contains := slices.Contains(AllowedMailDomains, domain)
	if !contains {
		return ErrEmailDomainNotAllowed.Error()
	}
	return
}

// checkCookieExpiry checks if the user's session cookie has expired.
// If the cookie has expired, it logs out the user, writes an Unauthorized response, and returns an error.
// It returns nil if the cookie is valid and has not expired.
func checkCookieExpiry(w http.ResponseWriter, r *http.Request) (err error) {
	username, err := UserState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrInvalidCookie.WriteJSON(w)
		return
	}

	cookieExpiry, err := UserState.Users().Get(username, "cookie-expiry")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrInvalidCookie.WriteJSON(w)
		return
	}

	cookieExpiryTime, err := time.Parse(time.UnixDate, cookieExpiry)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	if time.Now().After(cookieExpiryTime) {
		if UserState.IsLoggedIn(username) {
			UserState.Logout(username)
		}
		w.WriteHeader(http.StatusUnauthorized)
		ErrExpiredCookie.WriteJSON(w)
		return ErrExpiredCookie.Error()
	}

	return
}

// checkConfirmedMiddleware is a middleware that checks if the user is confirmed.
// If the user is not confirmed, it writes a Forbidden response.
// It calls the next handler if the user is confirmed.
func checkConfirmedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := UserState.UsernameCookie(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			ErrInvalidCookie.WriteJSON(w)
			return
		}

		if !UserState.IsConfirmed(username) {
			w.WriteHeader(http.StatusUnauthorized)
			ErrNotConfirmed.WriteJSON(w)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// checkCookieExpiryMiddleware is a middleware that checks if the user's session cookie has expired.
// If the cookie has expired, it writes an Unauthorized response and returns.
// It calls the next handler if the cookie is valid and has not expired.
func checkCookieExpiryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if expired := checkCookieExpiry(w, r); expired == nil {
			next.ServeHTTP(w, r)
		}
	}
}

// checkUserAlreadyGradedMiddleware is a middleware that checks if the user has already graded a course.
// It checks if the user has a grade for the specified course.
// If the user has already graded the course, it writes a Forbidden response and returns.
// It calls the next handler if the user has not yet graded the course and sets an empty grade for the user and course.
func checkUserAlreadyGradedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := UserState.UsernameCookie(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			ErrInvalidCookie.WriteJSON(w)
			return
		}

		courseCode := r.FormValue("code")
		if err = isEmptyStr(w, courseCode); err != nil {
			return
		}

		if _, err = UserState.Users().Get(username, courseCode); err == nil {
			w.WriteHeader(http.StatusForbidden)
			ErrCourseGraded.WriteJSON(w)
			return
		}

		next.ServeHTTP(w, r)

		UserState.Users().Set(username, courseCode, "")
	}
}
