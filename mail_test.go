package itpg

import (
	"testing"
)

func TestInitCreds(t *testing.T) {
	if err := initCreds("test.env"); err != nil {
		t.Error(err)
	}
}

func TestSendMail(t *testing.T) {
	if err := initCreds("test.env"); err != nil {
		t.Fatal(err)
	}
	if err := SendMail("Takumi Fujiwara", "takumi@fuji.ae", "iamsuperduperfastondownhills"); err != nil {
		t.Error(err)
	}
}
