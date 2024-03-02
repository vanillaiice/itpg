package itpg

import (
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
		ErrInvalidEmail.WriteJSON(w)
		return
	}
	if err = checkDomainAllowed(domain); err != nil {
		w.WriteHeader(http.StatusForbidden)
		ErrEmailDomainNotAllowed.WriteJSON(w)
		return
	}
	if UserState.HasUser(creds.Email) {
		if UserState.IsConfirmed(creds.Email) {
			w.WriteHeader(http.StatusForbidden)
			ErrUsernameTaken.WriteJSON(w)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			if UserState.CorrectPassword(creds.Email, creds.Password) {
				ErrNotConfirmed.WriteJSON(w)
			} else {
				ErrWrongUsernamePassword.WriteJSON(w)
			}
		}
		return
	}

	confirmationCode, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrGenCode.WriteJSON(w)
		return
	}

	if err = SendMailFunc(creds.Email, creds.Email, confirmationCode.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrSendMail.WriteJSON(w)
		return
	}

	UserState.AddUser(creds.Email, creds.Password, "")
	UserState.AddUnconfirmed(creds.Email, confirmationCode.String())

	Success.WriteJSON(w)
}

// SendNewConfirmationCode sends a new confirmation code to a registered user's email
// for confirmation.
func SendNewConfirmationCode(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	if !UserState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if UserState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		ErrConfirmed.WriteJSON(w)
		return
	}

	confirmationCode, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrGenCode.WriteJSON(w)
		return
	}
	if err = SendMailFunc(creds.Email, creds.Email, confirmationCode.String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrSendMail.WriteJSON(w)
		return
	}

	UserState.AddUnconfirmed(creds.Email, confirmationCode.String())

	Success.WriteJSON(w)
}

// Confirm confirms the user registration with the provided confirmation code.
func Confirm(w http.ResponseWriter, r *http.Request) {
	confirmationCode := r.FormValue("code")
	if err := isEmptyStr(w, confirmationCode); err != nil {
		return
	}

	if err := UserState.ConfirmUserByConfirmationCode(confirmationCode); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrWrongConfirmationCode.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// Login handles user login by checking credentials, confirming registration, setting a cookie
// with an expiry time, and logging the user in.
func Login(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}

	if !UserState.HasUser(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		ErrNotRegistered.WriteJSON(w)
		return
	}
	if !UserState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if !UserState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusUnauthorized)
		ErrNotConfirmed.WriteJSON(w)
		return
	}

	if err = UserState.Users().Set(creds.Email, "cookie-expiry", time.Now().Add(CookieTimeout).Format(time.UnixDate)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	if err = UserState.Login(w, creds.Email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// Logout logs out the currently logged-in user by removing their session.
func Logout(w http.ResponseWriter, r *http.Request) {
	username, err := UserState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrInvalidCookie.WriteJSON(w)
		return
	}

	if !UserState.IsLoggedIn(username) {
		w.WriteHeader(http.StatusForbidden)
		ErrNotLoggedIn.WriteJSON(w)
		return
	}

	UserState.Logout(username)

	Success.WriteJSON(w)
}

// ClearCookie clears the cookie for the current user session.
func ClearCookie(w http.ResponseWriter, r *http.Request) {
	UserState.ClearCookie(w)
	Success.WriteJSON(w)
}

// RefreshCookie refreshes the cookie for the current user session by updating its expiry time.
func RefreshCookie(w http.ResponseWriter, r *http.Request) {
	username, err := UserState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrInvalidCookie.WriteJSON(w)
		return
	}

	if err = UserState.Users().Set(username, "cookie-expiry", time.Now().Add(CookieTimeout).Format(time.UnixDate)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	if err = UserState.Login(w, username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	Success.WriteJSON(w)
}

// ChangePassword changes the account password of a currently logged-in user.
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	username, err := UserState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrInvalidCookie.WriteJSON(w)
		return
	}

	if !UserState.IsConfirmed(username) {
		w.WriteHeader(http.StatusForbidden)
		ErrNotConfirmed.WriteJSON(w)
		return
	}

	oldPassword, newPassword := r.FormValue("oldPassword"), r.FormValue("newPassword")
	if err = isEmptyStr(w, oldPassword, newPassword); err != nil {
		return
	}

	if !UserState.CorrectPassword(username, oldPassword) {
		w.WriteHeader(http.StatusUnauthorized)
		ErrWrongUsernamePassword.WriteJSON(w)
		return
	}

	UserState.SetPassword(username, newPassword)

	Success.WriteJSON(w)
}

// ResetPassword resets the password of an account, in case a user forgot it.
// func ResetPassword(w http.ResponseWriter, r *http.Request) {}

// DeleteAccount deletes the account of the currently logged-in user.
func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	username, err := UserState.UsernameCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		ErrInvalidCookie.WriteJSON(w)
		return
	}

	if err = UserState.Users().DelKey(username, "cookie-expiry"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrInternal.WriteJSON(w)
		return
	}

	UserState.RemoveUser(username)

	Success.WriteJSON(w)
}

// Ping checks that the user is logged in and that the cookie is not expired.
func Ping(w http.ResponseWriter, r *http.Request) {}
