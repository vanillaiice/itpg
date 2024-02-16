package itpg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/xyproto/permissionbolt/v2"
)

var creds = &Credentials{Email: "joe@joe.com", Password: "joejoejoe"}

func TestIsEmptyStr(t *testing.T) {
	w := httptest.NewRecorder()
	err := isEmptyStr(w, "foo", "bar", "baz")
	if err != nil {
		t.Error(err)
	}
	err = isEmptyStr(w, "", "bar")
	if err == nil {
		t.Error("expected failure")
	}
}

func TestDecodeCredentials(t *testing.T) {
	cb, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", bytes.NewReader(cb))
	if err != nil {
		t.Fatal(err)
	}
	creds_, err := decodeCredentials(w, r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(creds_, creds) {
		t.Errorf("got %v, want %v", creds_, creds)
	}
}

func TestExtractDomain(t *testing.T) {
	var err error
	email := "foo@bar.com"
	domain, err := extractDomain(email)
	if err != nil {
		t.Error(err)
	}
	if domain != "bar.com" {
		t.Errorf("got %s, want %s", domain, "bar.com")
	}
}

func TestValidAllowedDomains(t *testing.T) {
	var err error
	domains := []string{"foo.com", "bar.xyz", "buzz.io"}
	if err = validAllowedDomains(domains); err != nil {
		t.Error(err)
	}
	domains = []string{"*"}
	if err = validAllowedDomains(domains); err != nil {
		t.Error(err)
	}
	domains = []string{}
	if err = validAllowedDomains(domains); err == nil {
		t.Error("expected failure")
	}
}

func TestCheckDomainAllowed(t *testing.T) {
	var err error
	AllowedMailDomains = []string{"foo.com", "bar.xyz"}
	if err = checkDomainAllowed("foo.com"); err != nil {
		t.Error(err)
	}
	if err = checkDomainAllowed("foo.com"); err != nil {
		t.Error(err)
	}
	if err = checkDomainAllowed("buzz.cc"); err == nil {
		t.Error("expected failure")
	}
	AllowedMailDomains = []string{"*"}
	if err = checkDomainAllowed("fizz.cc"); err != nil {
		t.Error(err)
	}
}

func initTestUserState() (err error) {
	perm, err := permissionbolt.NewWithConf("userstate-test.db")
	if err != nil {
		return
	}
	userState = perm.UserState()
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

	Login(w, r)

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	err = checkCookieExpiry(w, r)
	if err != nil {
		t.Error(err)
	}

	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	c = &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value + "a",
	}
	r.AddCookie(c)

	err = checkCookieExpiry(w, r)
	if err == nil {
		t.Error("expected failure")
	}

	userState.Users().Set(
		creds.Email,
		"cookie-expiry",
		time.Now().Add(-time.Hour).Format(time.UnixDate),
	)

	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.AddCookie(c)

	err = checkCookieExpiry(w, r)
	if err == nil {
		t.Errorf("got %v, want %v", err, errExpiredCookie.Error())
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
	userState.Login(w, creds.Email)

	cookie := w.Result().Cookies()[0]
	c := &http.Cookie{
		Name:  cookie.Name,
		Value: cookie.Value,
	}
	r.AddCookie(c)

	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %v, want %v", w.Code, http.StatusUnauthorized)
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
	userState.Login(w, creds.Email)

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

	userState.Users().Set(
		creds.Email,
		"cookie-expiry",
		time.Now().Add(-time.Hour).Format(time.UnixDate),
	)

	r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	r.AddCookie(c)

	err = checkCookieExpiry(w, r)
	if err == nil {
		t.Errorf("got %v, want %v", err, errExpiredCookie.Error())
	}
}

func TestCheckUserAlreadyGradedMiddleware(t *testing.T) {
	err := initTestUserState()
	if err != nil {
		t.Error(err)
	}
	defer removeUserState()

	handler := func(w http.ResponseWriter, r *http.Request) {}
	middleware := checkUserAlreadyGradedMiddleware(handler)

	body, _ := json.Marshal(creds)
	r := httptest.NewRequest(http.MethodPost, "/grade?code="+courses[0].Code, bytes.NewReader(body))
	w := httptest.NewRecorder()

	userState.AddUser(creds.Email, creds.Password, "")
	userState.Confirm(creds.Email)

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

	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("got %v, want %v", w.Code, http.StatusForbidden)
	}
}
