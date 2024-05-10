package itpg

import (
	"fmt"
	"itpg/responses"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	zxcvbn "github.com/trustelem/zxcvbn"
)

// ConfirmationCodeValidityTime is the time during which the confimatoin code is valid.
const ConfirmationCodeValidityTime = time.Hour * 3

// KeyConfirmationCodeValidityTime is the key for geting the confirmation code validity time.
const KeyConfirmationCodeValidityTime = "cc_validity"

// MinPasswordScore is the minimum acceptable score of a password computed by zxcvbn.
const MinPasswordScore = 3

// CodeLength is the length of generated confirmation or reset code.
// The code is truncated from the beginning v4 uuid.
const CodeLength = 8

// Credentials represents the user credentials.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CredentialsReset represents the user credentials for resetting password.
type CredentialsReset struct {
	Code     string `json:"code"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CredentialsChange represents the user credentials for changing passwords.
type CredentialsChange struct {
	OldPassword string `json:"old"`
	NewPassword string `json:"new"`
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
		responses.ErrInvalidEmail.WriteJSON(w)
		return
	}
	if err = checkDomainAllowed(domain); err != nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrEmailDomainNotAllowed.WriteJSON(w)
		return
	}

	if score := zxcvbn.PasswordStrength(creds.Password, []string{}); score.Score < MinPasswordScore {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrWeakPassword.WriteJSON(w)
		return
	}

	if UserState.HasUser(creds.Email) {
		if UserState.IsConfirmed(creds.Email) {
			w.WriteHeader(http.StatusForbidden)
			responses.ErrRegistered.WriteJSON(w)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			if UserState.CorrectPassword(creds.Email, creds.Password) {
				responses.ErrNotConfirmed.WriteJSON(w)
			} else {
				responses.ErrWrongUsernamePassword.WriteJSON(w)
			}
		}
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrGenCode.WriteJSON(w)
		return
	}
	confirmationCode := uuid.String()[:CodeLength]

	if err = SendMailFunc(creds.Email, makeConfCodeMessage(creds.Email, confirmationCode)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrSendMail.WriteJSON(w)
		return
	}

	UserState.AddUser(creds.Email, creds.Password, "")
	UserState.AddUnconfirmed(creds.Email, confirmationCode)

	if err = UserState.Users().Set(creds.Email, KeyConfirmationCodeValidityTime, time.Now().Add(ConfirmationCodeValidityTime).String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
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
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if UserState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrConfirmed.WriteJSON(w)
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrGenCode.WriteJSON(w)
		return
	}
	confirmationCode := uuid.String()[:CodeLength]

	if err = SendMailFunc(creds.Email, makeConfCodeMessage(creds.Email, confirmationCode)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrSendMail.WriteJSON(w)
		return
	}

	UserState.AddUnconfirmed(creds.Email, confirmationCode)

	if err = UserState.Users().Set(creds.Email, KeyConfirmationCodeValidityTime, time.Now().Add(ConfirmationCodeValidityTime).String()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// Confirm confirms the user registration with the provided confirmation code.
func Confirm(w http.ResponseWriter, r *http.Request) {
	confirmationCode := r.FormValue("code")
	if err := isEmptyStr(w, confirmationCode); err != nil {
		return
	}

	username, err := UserState.FindUserByConfirmationCode(confirmationCode)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrNotRegistered.WriteJSON(w)
		return
	}

	confirmationCodeValidityTime, err := UserState.Users().Get(username, KeyConfirmationCodeValidityTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}
	t, err := time.Parse(time.RFC3339, confirmationCodeValidityTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}
	if t.After(time.Now()) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrConfirmationCodeExpired.WriteJSON(w)
		return
	}

	if err := UserState.ConfirmUserByConfirmationCode(confirmationCode); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongConfirmationCode.WriteJSON(w)
		return
	}

	if err := UserState.Users().DelKey(username, KeyConfirmationCodeValidityTime); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	UserState.RemoveUnconfirmed(username)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
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
		responses.ErrNotRegistered.WriteJSON(w)
		return
	}
	if !UserState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if !UserState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrNotConfirmed.WriteJSON(w)
		return
	}

	if err = UserState.Users().Set(creds.Email, CookieExpiryUserStateKey, time.Now().Add(CookieTimeout).Format(time.UnixDate)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if err = UserState.Login(w, creds.Email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// Logout logs out the currently logged-in user by removing their session.
func Logout(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(UsernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if !UserState.IsLoggedIn(username) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotLoggedIn.WriteJSON(w)
		return
	}

	UserState.Logout(username)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// ClearCookie clears the cookie for the current user session.
func ClearCookie(w http.ResponseWriter, r *http.Request) {
	UserState.ClearCookie(w)
	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// RefreshCookie refreshes the cookie for the current user session by updating its expiry time.
func RefreshCookie(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(UsernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if err := UserState.Users().Set(username, CookieExpiryUserStateKey, time.Now().Add(CookieTimeout).Format(time.UnixDate)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if err := UserState.Login(w, username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// ChangePassword changes the account password of a currently logged-in user.
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(UsernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if !UserState.IsConfirmed(username) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotConfirmed.WriteJSON(w)
		return
	}

	credsChange, err := decodeCredentialsChange(w, r)
	if err != nil {
		return
	}
	if err = isEmptyStr(w, credsChange.OldPassword, credsChange.NewPassword); err != nil {
		return
	}

	if !UserState.CorrectPassword(username, credsChange.OldPassword) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}

	UserState.SetPassword(username, credsChange.NewPassword)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// ResetPassword resets the account password of a user, in case it was forgotten.
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	credsReset, err := decodeCredentialsReset(w, r)
	if err != nil {
		return
	}

	var expectedResetCode string
	if expectedResetCode, err = UserState.Users().Get(credsReset.Email, "reset-code"); err != nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrResetCodeNotSent.WriteJSON(w)
		return
	}

	if credsReset.Code != expectedResetCode {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongResetCode.WriteJSON(w)
		return
	}

	if err = UserState.Users().DelKey(credsReset.Email, "reset-code"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	UserState.SetPassword(credsReset.Email, credsReset.Password)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// SendResetLink sends a mail containing a password reset link
func SendResetLink(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("email")
	if err := isEmptyStr(w, username); err != nil {
		return
	}

	if !UserState.HasUser(username) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotRegistered.WriteJSON(w)
		return
	}

	if _, err := UserState.Users().Get(username, "reset-code"); err == nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrResetCodeSent.WriteJSON(w)
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrGenCode.WriteJSON(w)
		return
	}
	resetCode := uuid.String()

	if err = SendMailFunc(username, makeResetCodeMessage(username, fmt.Sprintf("%s?code=%s", PasswordResetWebsiteURL, resetCode))); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrSendMail.WriteJSON(w)
		return
	}

	if err = UserState.Users().Set(username, "reset-code", resetCode); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// DeleteAccount deletes the account of the currently logged-in user.
func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		return
	}
	if !UserState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if !UserState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrNotConfirmed.WriteJSON(w)
		return
	}

	if err := UserState.Users().DelKey(creds.Email, CookieExpiryUserStateKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	UserState.RemoveUser(creds.Email)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// Ping checks that the user is logged in and that the cookie is not expired.
func Ping(w http.ResponseWriter, r *http.Request) {}
