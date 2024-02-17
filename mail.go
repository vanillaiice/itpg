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
	// SMTPHost is the host used for SMTP communication
	SMTPHost string
	// SMTPPort is the port number used for SMTP communication.
	SMTPPort string
	// SMTPURL is the full URL of the SMTP server, including the protocol and any additional path.
	SMTPURL string
	// MailFrom is the email address used as the sender in outgoing emails.
	MailFrom string
	// Username is the username used for authentication with the SMTP server.
	Username string
	// Password is the password used for authentication with the SMTP server.
	Password string
)

// InitCreds initializes SMTP credentials from the environment variables defined
// in the provided .env file path.
func InitCreds(envPath string) (err error) {
	if err = godotenv.Load(envPath); err != nil {
		return err
	}
	SMTPHost = os.Getenv("SMTP_HOST")
	SMTPPort = os.Getenv("SMTP_PORT")
	MailFrom = os.Getenv("MAIL_FROM")
	Username = os.Getenv("USERNAME")
	Password = os.Getenv("PASSWORD")
	for _, s := range []string{SMTPHost, SMTPPort, MailFrom, Username, Password} {
		if s == "" {
			return NewErrEmptyValueFor(s).Error()
		}
	}
	SMTPURL = fmt.Sprintf("%s:%s", SMTPHost, SMTPPort)

	return
}

// SendMail sends an email to the specified recipient containing the provided
// confirmation code for user registration.
func SendMail(mailToUsername, mailToAddress, confirmationCode string) error {
	auth := smtp.PlainAuth("", Username, Password, SMTPHost)

	message := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Confirmation Code\r\n\r\nHello %s,\r\n\nYour confirmation code: %s\r\n\nUse this code to complete your registration on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, MailFrom, time.Now().Format(time.RFC1123Z), mailToUsername, confirmationCode))

	return smtp.SendMail(SMTPURL, auth, MailFrom, []string{mailToAddress}, message)
}
