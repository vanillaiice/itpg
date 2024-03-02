package itpg

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// SMTP server configuration details
var (
	// MailSendFunc is the function used to send mails.
	SendMailFunc func(mailToUsername, mailToAddress, confirmationCode string) error
	// SMTPHost is the host used for SMTP communication.
	SMTPHost string
	// SMTPPort is the port number used for SMTP communication.
	SMTPPort string
	// SMTPURL is the full URL of the SMTP server, including the protocol and any additional path.
	SMTPURL string
	// MailFromAddress is the email address used as the sender in outgoing emails.
	MailFromAddress string
	// Username is the username used for authentication with the SMTP server.
	Username string
	// Password is the password used for authentication with the SMTP server.
	Password string
)

// InitCredsSMTP initializes SMTP credentials from the environment variables defined
// in the provided .env file path.
func InitCredsSMTP(envPath string, SMTPS bool) (err error) {
	if err = godotenv.Load(envPath); err != nil {
		return err
	}

	keys := []*string{&SMTPHost, &SMTPHost, &MailFromAddress}

	SendMailFunc = SendMailSMTP

	if SMTPS {
		SendMailFunc = SendMailSMTPS
		Username = os.Getenv("USERNAME")
		Password = os.Getenv("PASSWORD")
		keys = append(keys, &Username, &Password)
	}

	SMTPHost = os.Getenv("SMTP_HOST")
	SMTPPort = os.Getenv("SMTP_PORT")
	MailFromAddress = os.Getenv("MAIL_FROM")

	for _, s := range keys {
		if *s == "" {
			return ErrEmptyValue.Error()
		}
	}
	SMTPURL = fmt.Sprintf("%s:%s", SMTPHost, SMTPPort)

	return
}

// SendMailSMTPS sends an email using SMTP over TLS, with SMTP authentication.
func SendMailSMTPS(mailToUsername, mailToAddress, confirmationCode string) error {
	auth := smtp.PlainAuth("", Username, Password, SMTPHost)
	message := makeMessage(mailToAddress, mailToUsername, confirmationCode)
	return smtp.SendMail(SMTPURL, auth, MailFromAddress, []string{mailToAddress}, message)
}

// SendMailSMTP sends an email using SMTP without authentication.
// This should only be used when the SMTP server and the itpg-backend
// binary are running on the same machine.
func SendMailSMTP(mailToUsername, mailToAddress, confirmationCode string) error {
	c, err := smtp.Dial(SMTPURL)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.Mail(MailFromAddress); err != nil {
		return err
	}

	if err = c.Rcpt(mailToAddress); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	message := makeMessage(mailToAddress, mailToUsername, confirmationCode)
	if _, err := w.Write(message); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	if err = c.Quit(); err != nil {
		return err
	}

	return nil
}

// makeMessage creates the email message to be sent.
func makeMessage(mailToAddress, mailToUsername, confirmationCode string) []byte {
	return []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Confirmation Code\r\n\r\nHello %s,\r\n\nYour confirmation code: %s\r\n\nUse this code to complete your registration on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, MailFromAddress, time.Now().Format(time.RFC1123Z), mailToUsername, confirmationCode))
}
