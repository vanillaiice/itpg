package itpg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeCredentials(t *testing.T) {
	var c = &Credentials{Username: "foo", Password: "bar", Email: "foo@bar.baz"}
	cb, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "", bytes.NewReader(cb))
	if err != nil {
		t.Fatal(err)
	}
	creds, err := decodeCredentials(w, r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(creds, c) {
		t.Errorf("got %v, want %v", creds, c)
	}
}

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
