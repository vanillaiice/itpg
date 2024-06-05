package itpg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var creds = &Credentials{Email: "joe@joe.com", Password: "joejoejoe"}
var credsReset = &CredentialsReset{Email: "joe@joe.com", Password: "joejoejoe", Code: "mynameisjoe"}
var credsChange = &CredentialsChange{OldPassword: "joejoejoe", NewPassword: "eojeojeoj"}
var gradeData = &GradeData{CourseCode: "foo", ProfUUID: "bar", GradeTeaching: 5, GradeCoursework: 4, GradeLearning: 3}

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
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	s, err := decodeCredentials(w, r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(s, creds) {
		t.Errorf("got %v, want %v", s, creds)
	}
}

func TestDecodeCredentialsReset(t *testing.T) {
	b, err := json.Marshal(credsReset)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	s, err := decodeCredentialsReset(w, r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(s, credsReset) {
		t.Errorf("got %v, want %v", s, credsReset)
	}
}

func TestDecodeCredentialsChange(t *testing.T) {
	b, err := json.Marshal(credsChange)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	s, err := decodeCredentialsChange(w, r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(s, credsChange) {
		t.Errorf("got %v, want %v", s, credsChange)
	}
}

func TestDecodeGradeData(t *testing.T) {
	b, err := json.Marshal(gradeData)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	s, err := decodeGradeData(w, r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(s, gradeData) {
		t.Errorf("got %v, want %v", s, gradeData)
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
	allowedMailDomains = []string{"foo.com", "bar.xyz"}
	if err = checkDomainAllowed("foo.com"); err != nil {
		t.Error(err)
	}
	if err = checkDomainAllowed("foo.com"); err != nil {
		t.Error(err)
	}
	if err = checkDomainAllowed("buzz.cc"); err == nil {
		t.Error("expected failure")
	}
	allowedMailDomains = []string{"*"}
	if err = checkDomainAllowed("fizz.cc"); err != nil {
		t.Error(err)
	}
}
