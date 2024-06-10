package server

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
	userState = perm.UserState()
	cookieTimeout = 1 * time.Minute
	userState.SetCookieTimeout(int64(cookieTimeout.Seconds()))
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

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)

	w := httptest.NewRecorder()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	login(w, r)

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

	if err = userState.Users().Set(
		creds.Email,
		cookieExpiryUserStateKey,
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

	userState.AddUser(creds.Email, creds.Password, "")
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
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

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)

	login(w, r)

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

	if err = userState.Users().Set(
		creds.Email,
		cookieExpiryUserStateKey,
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

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got %v, want %v", w.Code, http.StatusOK)
	}
}

func TestCheckAdminMiddleware_NotAdmin(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	handler := func(w http.ResponseWriter, r *http.Request) {}

	middleware := checkAdminMiddleware(handler)

	w := httptest.NewRecorder()

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestCheckAdminMiddleware_Admin(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	handler := func(w http.ResponseWriter, r *http.Request) {}

	middleware := checkAdminMiddleware(handler)

	w := httptest.NewRecorder()

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)
	userState.SetAdminStatus(creds.Email)
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got %v, want %v", w.Code, http.StatusOK)
	}
}

func TestCheckSuperAdminMiddleware_User(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	handler := func(w http.ResponseWriter, r *http.Request) {}

	middleware := checkSuperAdminMiddleware(handler)

	w := httptest.NewRecorder()

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestCheckSuperAdminMiddleware_Admin(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	handler := func(w http.ResponseWriter, r *http.Request) {}

	middleware := checkSuperAdminMiddleware(handler)

	w := httptest.NewRecorder()

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)
	userState.SetAdminStatus(creds.Email)
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestCheckSuperAdminMiddleware_SuperAdmin(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	handler := func(w http.ResponseWriter, r *http.Request) {}

	middleware := checkSuperAdminMiddleware(handler)

	w := httptest.NewRecorder()

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)
	userState.SetAdminStatus(creds.Email)
	userState.SetBooleanField(creds.Email, "super", true)
	if err = userState.Login(w, creds.Email); err != nil {
		t.Fatal(err)
	}

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	r = r.WithContext(context.WithValue(r.Context(), usernameContextKey, creds.Email))
	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got %v, want %v", w.Code, http.StatusOK)
	}
}
