package itpg

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/vanillaiice/itpg/responses"
	"github.com/xyproto/permissionbolt/v2"
)

func initTestUserState() (err error) {
	perm, err := permissionbolt.NewWithConf("userstate-test.db")
	if err != nil {
		return
	}
	UserState = perm.UserState()
	CookieTimeout = 1 * time.Minute
	UserState.SetCookieTimeout(int64(CookieTimeout.Seconds()))
	return
}

func removeUserState() {
	_ = os.Remove("userstate-test.db")
}

func TestCheckCookieExpiry(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	UserState.AddUser(creds.Email, creds.Password, "")
	UserState.Confirm(creds.Email)

	w := httptest.NewRecorder()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	Login(w, r)

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	err = checkCookieExpiry(creds.Email)
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c = &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	err = checkCookieExpiry(creds.Email + "a")
	if err == nil {
		t.Error("expected failure")
	}

	if err = UserState.Users().Set(
		creds.Email,
		CookieExpiryUserStateKey,
		time.Now().Add(-time.Hour).Format(time.UnixDate),
	); err != nil {
		t.Fatal(err)
	}

	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.AddCookie(c)

	err = checkCookieExpiry(creds.Email)
	if err == nil {
		t.Errorf("got %v, want %v", err, responses.ErrExpiredCookie.Error())
	}
}

func TestCheckConfirmedMiddleware_Unconfirmed(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	handler := func(w http.ResponseWriter, r *http.Request) {}
	middleware := checkConfirmedMiddleware(handler)

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	UserState.AddUser(creds.Email, creds.Password, "")
	if err = UserState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), UsernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestCheckCookieExpiryMiddleware(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	handler := func(w http.ResponseWriter, r *http.Request) {}
	middleware := checkCookieExpiryMiddleware(handler)

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	UserState.AddUser(creds.Email, creds.Password, "")
	UserState.Confirm(creds.Email)

	Login(w, r)

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got %v, want %v", w.Code, http.StatusOK)
	}

	if err = UserState.Users().Set(
		creds.Email,
		CookieExpiryUserStateKey,
		time.Now().Add(-time.Hour).Format(time.UnixDate),
	); err != nil {
		t.Fatal(err)
	}

	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.AddCookie(c)

	err = checkCookieExpiry(creds.Email)
	if err == nil {
		t.Errorf("got %v, want %v", err, responses.ErrExpiredCookie.Error())
	}
}

func TestCheckConfirmedMiddleware_Confirmed(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	handler := func(w http.ResponseWriter, r *http.Request) {}
	middleware := checkConfirmedMiddleware(handler)

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	UserState.AddUser(creds.Email, creds.Password, "")
	UserState.Confirm(creds.Email)
	if err = UserState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), UsernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got %v, want %v", w.Code, http.StatusOK)
	}
}
