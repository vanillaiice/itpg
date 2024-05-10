package responses

import (
	"encoding/json"
	"fmt"
	"io"
)

// Response represents a response returned by the server.
type Response struct {
	Code    int         `json:"code"`    // Internal response status code
	Message interface{} `json:"message"` // Message associated with the response
}

// Error returns an error representation of the Response.
// It contains the error code and the message.
func (r *Response) Error() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// WriteJSON writes a response to the specified writer
func (r *Response) WriteJSON(w io.Writer) {
	w.Write([]byte(r.Error()))
}

// NewResponse creates a new Response with the given code and message.
func NewResponse(code int, message interface{}) *Response {
	return &Response{Code: code, Message: message}
}

// NewErrEmptyValueFor returns a new Response struct with an error code
// indicating an empty value, and the name of the empty value
func NewErrEmptyValueFor(s string) *Response {
	return &Response{ErrEmptyValue.Code, fmt.Sprintf("got empty value for %s", s)}
}

// SucessCode indicates a successful operation.
var SuccessCode = 2000

// Success response indicating successful operation.
var Success = NewResponse(SuccessCode, "success")

// Client-side errors
var (
	// ErrRegistered indicates that the user is already registered.
	ErrRegistered = NewResponse(4000, "already registered")
	// ErrNotRegistered indicates that the user is not registered.
	ErrNotRegistered = NewResponse(4001, "not registered")
	// ErrUsernameTaken indicates that the username is already taken.
	ErrUsernameTaken = NewResponse(4002, "username taken")
	// ErrLoggedIn indicates that the user is already logged in.
	ErrLoggedIn = NewResponse(4003, "already logged in")
	// ErrNotLoggedIn indicates that the user is not logged in.
	ErrNotLoggedIn = NewResponse(4004, "not logged in")
	// ErrConfirmed indicates that the user account is already confirmed.
	ErrConfirmed = NewResponse(4005, "already confirmed")
	// ErrNotConfirmed indicates that the user account is not confirmed.
	ErrNotConfirmed = NewResponse(4006, "not confirmed")
	// ErrNotConfirmedUser indicates that the user is not confirmed.
	ErrNotConfirmedUser = NewResponse(4007, "user not confirmed")
	// ErrWrongUsernamePassword indicates that the username or password is incorrect.
	ErrWrongUsernamePassword = NewResponse(4008, "wrong username or password")
	// ErrWrongConfirmationCode indicates that the confirmation code is incorrect.
	ErrWrongConfirmationCode = NewResponse(4009, "wrong confirmation code")
	// ErrInvalidCookie indicates that the cookie is invalid.
	ErrInvalidCookie = NewResponse(4010, "invalid cookie")
	// ErrExpiredCookie indicates that the cookie has expired.
	ErrExpiredCookie = NewResponse(4011, "expired cookie")
	// ErrBadRequest indicates a bad request.
	ErrBadRequest = NewResponse(4012, "bad request")
	// ErrEmptyValue indicates an empty value.
	ErrEmptyValue = NewResponse(4013, "got empty value")
	// ErrCourseGraded indicates that the course is already graded.
	ErrCourseGraded = NewResponse(4014, "course already graded")
	// ErrPermissionDenied indicates that permission is denied.
	ErrPermissionDenied = NewResponse(4015, "permission denied")
	// ErrInvalidEmail indicates that the email format is invalid.
	ErrInvalidEmail = NewResponse(4016, "invalid email format")
	// ErrEmailDomainNotAllowed indicates that the email domain is not allowed.
	ErrEmailDomainNotAllowed = NewResponse(4017, "email domain not allowed")
	// ErrRequestLimitReached indicates that the user has reached the request limit
	ErrRequestLimitReached = NewResponse(4018, "request limit reached")
	// ErrResetCodeSent indicates that a reset code was already sent
	ErrResetCodeSent = NewResponse(4019, "reset code already sent")
	// ErrResetCodeNotSent indicates that a reset code was not sent
	ErrResetCodeNotSent = NewResponse(4019, "reset code not sent")
	// ErrWrongResetCode indicates that the reset code is incorrect.
	ErrWrongResetCode = NewResponse(4020, "wrong reset code")
	// ErrWeakPassword indicates that the provided password is weak.
	ErrWeakPassword = NewResponse(4021, "weak password")
	// ErrConfirmationCodeExpired indicates that the provided confirmation code is expired.
	ErrConfirmationCodeExpired = NewResponse(4022, "confirmation code expired")
)

// Server-side Errors
var (
	// ErrGenCode indicates an error generating the confirmation code.
	ErrGenCode = NewResponse(5000, "error generating confirmation code")
	// ErrSendMail indicates an error mailing the confirmation code.
	ErrSendMail = NewResponse(5001, "error mailing confirmation code")
	// ErrInternal indicates an internal Error.
	ErrInternal = NewResponse(5002, "internal error")
)
