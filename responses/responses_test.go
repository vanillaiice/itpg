package responses

import (
	"bytes"
	"testing"
)

func TestNewResponse(t *testing.T) {
	code := 1234
	message := "test message"
	resp := NewResponse(code, message)
	if resp.Code != code || resp.Message != message {
		t.Errorf("expected %d, %s, got %d, %v", code, message, resp.Code, resp.Message)
	}
}

func TestNewErrEmptyValueFor(t *testing.T) {
	fieldName := "test_field"
	resp := NewErrEmptyValueFor(fieldName)
	expected := ErrEmptyValue.Code
	if resp.Code != expected {
		t.Errorf("expected %d, got %d", expected, resp.Code)
	}
	expectedMessage := "got empty value for test_field"
	if resp.Message != expectedMessage {
		t.Errorf("expected %s, got %v", expectedMessage, resp.Message)
	}
}

func TestResponseError(t *testing.T) {
	code := 5000
	message := "test error"
	resp := NewResponse(code, message)
	expected := `{"code":5000,"message":"test error"}`
	if resp.Error() != expected {
		t.Errorf("got %s, want %s", resp.Error(), expected)
	}
}

func TestResponseWriteJSON(t *testing.T) {
	code := 4000
	message := "test write"
	resp := NewResponse(code, message)
	buf := new(bytes.Buffer)
	resp.WriteJSON(buf)
	expected := `{"code":4000,"message":"test write"}`
	if buf.String() != expected {
		t.Errorf(" got %s, want %s", buf.String(), expected)
	}
}

func TestErrorCodes(t *testing.T) {
	// Test client-side errors
	testErrorCodes(t, []struct {
		err  *Response
		code int
	}{
		{ErrRegistered, 4000},
		{ErrNotRegistered, 4001},
		{ErrUsernameTaken, 4002},
		{ErrLoggedIn, 4003},
		{ErrNotLoggedIn, 4004},
		{ErrConfirmed, 4005},
		{ErrNotConfirmed, 4006},
		{ErrNotConfirmedUser, 4007},
		{ErrWrongUsernamePassword, 4008},
		{ErrWrongConfirmationCode, 4009},
		{ErrInvalidCookie, 4010},
		{ErrExpiredCookie, 4011},
		{ErrBadRequest, 4012},
		{ErrEmptyValue, 4013},
		{ErrCourseGraded, 4014},
		{ErrPermissionDenied, 4015},
		{ErrInvalidEmail, 4016},
		{ErrEmailDomainNotAllowed, 4017},
		{ErrRequestLimitReached, 4018},
		{ErrResetCodeSent, 4019},
		{ErrResetCodeNotSent, 4019},
		{ErrWrongResetCode, 4020},
		{ErrWeakPassword, 4021},
	})

	// Test server-side errors
	testErrorCodes(t, []struct {
		err  *Response
		code int
	}{
		{ErrGenCode, 5000},
		{ErrSendMail, 5001},
		{ErrInternal, 5002},
	})
}

func testErrorCodes(t *testing.T, testCases []struct {
	err  *Response
	code int
}) {
	for _, tc := range testCases {
		if tc.err.Code != tc.code {
			t.Errorf("%v: expected %d, got %d", tc.err.Message, tc.code, tc.err.Code)
		}
	}
}
