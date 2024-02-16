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
	// smtpHost is the host used for SMTP communication
	smtpHost string
	// smtpPort is the port number used for SMTP communication.
	smtpPort string
	// smtpURL is the full URL of the SMTP server, including the protocol and any additional path.
	smtpURL string
	// mailFrom is the email address used as the sender in outgoing emails.
	mailFrom string
	// username is the username used for authentication with the SMTP server.
	username string
	// password is the password used for authentication with the SMTP server.
	password string
)

// initCreds initializes SMTP credentials from the environment variables defined
// in the provided .env file path.
func initCreds(envPath string) (err error) {
	if err = godotenv.Load(envPath); err != nil {
		return err
	}
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	mailFrom = os.Getenv("MAIL_FROM")
	username = os.Getenv("USERNAME")
	password = os.Getenv("PASSWORD")
	for _, s := range []string{smtpHost, smtpPort, mailFrom, username, password} {
		if s == "" {
			return NewErrEmptyValueFor(s).Error()
		}
	}
	smtpURL = fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return
}

// SendMail sends an email to the specified recipient containing the provided
// confirmation code for user registration.
func SendMail(usernameMailTo, mailToAddress, confirmationCode string) error {
	auth := smtp.PlainAuth("", username, password, smtpHost)

	message := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Confirmation Code\r\n\r\nHello %s,\r\n\nYour confirmation code: %s\r\n\nUse this code to complete your registration on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, mailFrom, time.Now().Format(time.RFC1123Z), usernameMailTo, confirmationCode))

	return smtp.SendMail(smtpURL, auth, mailFrom, []string{mailToAddress}, message)
}
