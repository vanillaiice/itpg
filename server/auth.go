package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"
	"github.com/trustelem/zxcvbn"
	"github.com/vanillaiice/itpg/responses"
)

// keyConfirmationCodeValidityTime is the key for geting the confirmation code validity time.
const keyConfirmationCodeValidityTime = "cc_validity"

// confirmationCodeValidityTime is the time during which the confimatoin code is valid.
var confirmationCodeValidityTime time.Duration

// minPasswordScore is the minimum acceptable score of a password computed by zxcvbn.
var minPasswordScore int

// codeLength is the length of generated codes.
// The code is truncated from the beginning of a v4 uuid.
var codeLength int

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

// allowedMailDomains are the email domains allowed to register.
// If the first item of the slice is "*", all domains will be allowed.
var allowedMailDomains []string

// register handles user registration by validating credentials, generating a confirmation
// code, sending an email with the code, and adding the user to the system.
func register(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	domain, err := extractDomain(creds.Email)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrInvalidEmail.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}
	if err = checkDomainAllowed(domain); err != nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrEmailDomainNotAllowed.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	if score := zxcvbn.PasswordStrength(creds.Password, []string{}); score.Score < minPasswordScore {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrWeakPassword.WriteJSON(w)
		return
	}

	if userState.HasUser(creds.Email) {
		if userState.IsConfirmed(creds.Email) {
			w.WriteHeader(http.StatusForbidden)
			responses.ErrRegistered.WriteJSON(w)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			if userState.CorrectPassword(creds.Email, creds.Password) {
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
		log.Error().Msg(err.Error())
		return
	}
	confirmationCode := uuid.String()[:codeLength]

	if err = mailer.SendMail(creds.Email, mailer.MakeConfCodeMessage(creds.Email, confirmationCode)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrSendMail.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	userState.AddUser(creds.Email, creds.Password, "")
	userState.AddUnconfirmed(creds.Email, confirmationCode)

	if err = userState.Users().Set(creds.Email, keyConfirmationCodeValidityTime, time.Now().Add(confirmationCodeValidityTime).Format(time.RFC3339)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// sendNewConfirmationCode sends a new confirmation code to a registered user's email
// for confirmation.
func sendNewConfirmationCode(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	if !userState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if userState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrConfirmed.WriteJSON(w)
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrGenCode.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}
	confirmationCode := uuid.String()[:codeLength]

	if err = mailer.SendMail(creds.Email, mailer.MakeConfCodeMessage(creds.Email, confirmationCode)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrSendMail.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	userState.AddUnconfirmed(creds.Email, confirmationCode)

	if err = userState.Users().Set(creds.Email, keyConfirmationCodeValidityTime, time.Now().Add(confirmationCodeValidityTime).Format(time.RFC3339)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// confirm confirms the user registration with the provided confirmation code.
func confirm(w http.ResponseWriter, r *http.Request) {
	confirmationCode := r.FormValue("code")
	if err := isEmptyStr(w, confirmationCode); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	username, err := userState.FindUserByConfirmationCode(confirmationCode)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrNotRegistered.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	confirmationCodeValidityTime, err := userState.Users().Get(username, keyConfirmationCodeValidityTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}
	t, err := time.Parse(time.RFC3339, confirmationCodeValidityTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}
	if !t.After(time.Now()) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrConfirmationCodeExpired.WriteJSON(w)
		return
	}

	if err := userState.ConfirmUserByConfirmationCode(confirmationCode); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongConfirmationCode.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	if err := userState.Users().DelKey(username, keyConfirmationCodeValidityTime); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	userState.RemoveUnconfirmed(username)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// login handles user login by checking credentials, confirming registration, setting a cookie
// with an expiry time, and logging the user in.
func login(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	if !userState.HasUser(creds.Email) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotRegistered.WriteJSON(w)
		return
	}
	if !userState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if !userState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrNotConfirmed.WriteJSON(w)
		return
	}

	if err = userState.Users().Set(creds.Email, cookieExpiryUserStateKey, time.Now().Add(cookieTimeout).Format(time.UnixDate)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	if err = userState.Login(w, creds.Email); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// logout logs out the currently logged-in user by removing their session.
func logout(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(usernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if !userState.IsLoggedIn(username) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotLoggedIn.WriteJSON(w)
		return
	}

	userState.Logout(username)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// clearCookie clears the cookie for the current user session.
func clearCookie(w http.ResponseWriter, r *http.Request) {
	userState.ClearCookie(w)
	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// refreshCookie refreshes the cookie for the current user session by updating its expiry time.
func refreshCookie(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(usernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if err := userState.Users().Set(username, cookieExpiryUserStateKey, time.Now().Add(cookieTimeout).Format(time.UnixDate)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	if err := userState.Login(w, username); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// changePassword changes the account password of a currently logged-in user.
func changePassword(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(usernameContextKey).(string)
	if !ok || username == "" {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		return
	}

	if !userState.IsConfirmed(username) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotConfirmed.WriteJSON(w)
		return
	}

	credsChange, err := decodeCredentialsChange(w, r)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	if err = isEmptyStr(w, credsChange.OldPassword, credsChange.NewPassword); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	if !userState.CorrectPassword(username, credsChange.OldPassword) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}

	userState.SetPassword(username, credsChange.NewPassword)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// resetPassword resets the account password of a user, in case it was forgotten.
func resetPassword(w http.ResponseWriter, r *http.Request) {
	credsReset, err := decodeCredentialsReset(w, r)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	var expectedResetCode string
	if expectedResetCode, err = userState.Users().Get(credsReset.Email, "reset-code"); err != nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrResetCodeNotSent.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	if credsReset.Code != expectedResetCode {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongResetCode.WriteJSON(w)
		return
	}

	if err = userState.Users().DelKey(credsReset.Email, "reset-code"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	userState.SetPassword(credsReset.Email, credsReset.Password)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// sendResetLink sends a mail containing a password reset link
func sendResetLink(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("email")
	if err := isEmptyStr(w, username); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	if !userState.HasUser(username) {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrNotRegistered.WriteJSON(w)
		return
	}

	if _, err := userState.Users().Get(username, "reset-code"); err == nil {
		w.WriteHeader(http.StatusForbidden)
		responses.ErrResetCodeSent.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrGenCode.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}
	resetCode := uuid.String()

	if err = mailer.SendMail(username, mailer.MakeResetCodeMessage(username, fmt.Sprintf("%s?code=%s", passwordResetUrl, resetCode))); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrSendMail.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	if err = userState.Users().Set(username, "reset-code", resetCode); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// deleteAccount deletes the account of the currently logged-in user.
func deleteAccount(w http.ResponseWriter, r *http.Request) {
	creds, err := decodeCredentials(w, r)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	if !userState.CorrectPassword(creds.Email, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrWrongUsernamePassword.WriteJSON(w)
		return
	}
	if !userState.IsConfirmed(creds.Email) {
		w.WriteHeader(http.StatusUnauthorized)
		responses.ErrNotConfirmed.WriteJSON(w)
		return
	}

	if err := userState.Users().DelKey(creds.Email, cookieExpiryUserStateKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responses.ErrInternal.WriteJSON(w)
		log.Error().Msg(err.Error())
		return
	}

	userState.RemoveUser(creds.Email)

	w.Header().Set("Content-Type", "application/json")
	responses.Success.WriteJSON(w)
}

// ping checks that the user is logged in and that the cookie is not expired.
func ping(w http.ResponseWriter, r *http.Request) {}
