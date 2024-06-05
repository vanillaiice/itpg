package mail

import (
	"fmt"
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
)

func initTestSmtpServer() (*smtpmock.Server, error) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	if err := server.Start(); err != nil {
		return nil, err
	}
	return server, nil
}

func TestInitCredsSmtp(t *testing.T) {
	_, err := NewClient(".env.test", false)
	if err != nil {
		t.Error(err)
	}
}

func TestInitCredsSmtps(t *testing.T) {
	_, err := NewClient(".env.test", true)
	if err != nil {
		t.Error(err)
	}
}

func TestSendMailSmtp(t *testing.T) {
	server, err := initTestSmtpServer()
	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("127.0.0.1:%d", server.PortNumber())
	mailFromAddress := "testing@test.com"

	if err := sendMailSmtp(url, mailFromAddress, "takumi@fuji.ae", []byte("iamsuperduperfastondownhills")); err != nil {
		t.Error(err)
	}

	if err = server.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestSendMailSmtps(t *testing.T) {
	server, err := initTestSmtpServer()
	if err != nil {
		t.Fatal(err)
	}

	/* The code block below will fail because the go-mock-smtp package does not support auth.
	if err := SendMailSMTPS("takumi@fuji.jp", "iamsuperduperfastondownhills"); err != nil {
		t.Error(err)
	}
	*/

	if err = server.Stop(); err != nil {
		t.Fatal(err)
	}
}
