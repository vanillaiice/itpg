package itpg

import (
	"fmt"
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
)

func initTestSMTPServer() (*smtpmock.Server, error) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	if err := server.Start(); err != nil {
		return nil, err
	}
	return server, nil
}

func TestInitCredsSMTP(t *testing.T) {
	if err := InitCredsSMTP("test.env", false); err != nil {
		t.Error(err)
	}
}

func TestInitCredsSMTPS(t *testing.T) {
	if err := InitCredsSMTP("test.env", true); err != nil {
		t.Error(err)
	}
}

func TestSendMailSMTP(t *testing.T) {
	server, err := initTestSMTPServer()
	if err != nil {
		t.Fatal(err)
	}

	SMTPURL = fmt.Sprintf("127.0.0.1:%d", server.PortNumber())
	MailFromAddress = "testing@test.com"

	if err := SendMailSMTP("takumi@fuji.ae", []byte("iamsuperduperfastondownhills")); err != nil {
		t.Error(err)
	}

	if err = server.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestSendMailSMTPS(t *testing.T) {
	server, err := initTestSMTPServer()
	if err != nil {
		t.Fatal(err)
	}

	Username = "tester"
	Password = "testtter"
	SMTPHost = "127.0.0.1"
	SMTPURL = fmt.Sprintf("%s:%d", SMTPHost, server.PortNumber())
	MailFromAddress = "testing@test.com"

	/* The below code block will fail, because the go-mock-smtp package does not support auth.
	if err := SendMailSMTPS("takumi@fuji.ae", "iamsuperduperfastondownhills"); err != nil {
		t.Error(err)
	}
	*/

	if err = server.Stop(); err != nil {
		t.Fatal(err)
	}
}
