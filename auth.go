package itpg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

// Credentials represents the user credentials for authentication and registration.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AllowedMailDomains are the email domains allowed to register.
// If the first item of the slice is "*", all domains will be allowed.
var AllowedMailDomains []string

// Register handles user registration by validating credentials, generating a confirmation
// code, sending an email with the code, and adding the user to the system.
func Register(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	domain, err := extractDomain(creds.Email)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		errInvalidEmail.WriteJSON(w)
		return
	}
	if err = checkDomainAllowed(domain); err != nil {
		w.WriteHeader(http.StatusForbidden)
		errEmailDomainNotAllowed.WriteJSON(w)
		return
	}
	if userState.HasUser(creds.Email) {
		if userState.IsConfirmed(creds.Email) {
			w.WriteHeader(http.StatusForbidden)
			errUsernameTaken.WriteJSON(w)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			if userState.CorrectPassword(creds.Email, creds.Password) {
				errNotConfirmed.WriteJSON(w)
			} else {
				errWrongUsernamePassword.WriteJSON(w)
			}
		}
		return
	}

	confirmationCode, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errGenCode.WriteJSON(w)
		return
	}
	if err = SendMail(creds.Email, creds.Email, confirmationCode.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errSendMail.WriteJSON(w)
		return
	}

	userState.AddUser(creds.Email, creds.Password, "")
	userState.AddUnconfirmed(creds.Email, confirmationCode.String())

	success.WriteJSON(w)
}

// SendNewConfirmationCode sends a new confirmation code to a registered user's email
// for confirmation.
func SendNewConfirmationCode(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	if !userState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		errWrongUsernamePassword.WriteJSON(w)
		return
	}
	if userState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		errConfirmed.WriteJSON(w)
		return
	}

	confirmationCode, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errGenCode.WriteJSON(w)
		return
	}
	if err = SendMail(creds.Email, creds.Email, confirmationCode.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errSendMail.WriteJSON(w)
		return
	}

	userState.AddUnconfirmed(creds.Email, confirmationCode.String())

	success.WriteJSON(w)
}

// Confirm confirms the user registration with the provided confirmation code.
func Confirm(w http.ResponseWriter, r *http.Request) {
	confirmationCode := r.FormValue("code")
	if err := isEmptyStr(w, confirmationCode); err != nil {
		return
	}

	if err := userState.ConfirmUserByConfirmationCode(confirmationCode); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errWrongConfirmationCode.WriteJSON(w)
		return
	}

	success.WriteJSON(w)
}

// Login handles user login by checking credentials, confirming registration, setting a cookie
// with an expiry time, and logging the user in.
func Login(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	if !userState.HasUser(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		errNotRegistered.WriteJSON(w)
		return
	}
	if !userState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		errWrongUsernamePassword.WriteJSON(w)
		return
	}
	if !userState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusUnauthorized)
		errNotConfirmed.WriteJSON(w)
		return
	}

	userState.Users().Set(
		creds.Email,
		"cookie-expiry",
		time.Now().Add(cookieTimeout).Format(time.UnixDate),
	)

	userState.Login(w, creds.Email)

	success.WriteJSON(w)
}

// Logout logs out the currently logged-in user by removing their session.
func Logout(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errInvalidCookie.WriteJSON(w)
		return
	}

	if !userState.IsLoggedIn(username) {
		w.WriteHeader(http.StatusForbidden)
		errNotLoggedIn.WriteJSON(w)
		return
	}

	userState.Logout(username)

	success.WriteJSON(w)
}

// ClearCookie clears the cookie for the current user session.
func ClearCookie(w http.ResponseWriter, r *http.Request) {
	userState.ClearCookie(w)
	success.WriteJSON(w)
}

// RefreshCookie refreshes the cookie for the current user session by updating its expiry time.
func RefreshCookie(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errInvalidCookie.WriteJSON(w)
		return
	}

	userState.Users().Set(
		username,
		"cookie-expiry",
		time.Now().Add(cookieTimeout).Format(time.UnixDate),
	)

	userState.Login(w, username)

	success.WriteJSON(w)
}

// DeleteAccount deletes the account of the currently logged-in user.
func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errInvalidCookie.WriteJSON(w)
		return
	}

	userState.Users().DelKey(username, "cookie-expiry")
	userState.RemoveUser(username)

	success.WriteJSON(w)
}

// Greet greets the user with a personalized message, including a cowboy emoji 🤠.
func Greet(w http.ResponseWriter, r *http.Request) {
	username, err := userState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errInvalidCookie.WriteJSON(w)
		return
	}
	(&Response{Code: successCode, Message: fmt.Sprintf("Sup %s 🤠 ?", username)}).WriteJSON(w)
}

// Ping checks that the user is logged in and that the cookie is not expired.
func Ping(w http.ResponseWriter, r *http.Request) {}
