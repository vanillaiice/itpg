package itpg

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/vanillaiice/itpg/responses"
)

// smtp server configuration details
var (
	// mailSendFunc is the function used to send mails.
	sendMailFunc func(mailToAddress string, message []byte) error
	// smtpHost is the host used for smtp communication.
	smtpHost string
	// smtpPort is the port number used for smtp communication.
	smtpPort string
	// smtpUrl is the full URL of the smtp server, including the protocol and any additional path.
	smtpUrl string
	// mailFromAddress is the email address used as the sender in outgoing emails.
	mailFromAddress string
	// username is the username used for authentication with the smtp server.
	username string
	// password is the password used for authentication with the smtp server.
	password string
)

// initCredsSmtp initializes smtp credentials from the environment variables defined
// in the provided .env file path.
func initCredsSmtp(envPath string, smtps bool) (err error) {
	if err = godotenv.Load(envPath); err != nil {
		return err
	}

	keys := []*string{&smtpHost, &smtpHost, &mailFromAddress}

	sendMailFunc = sendMailSmtp

	if smtps {
		sendMailFunc = sendMailSmtps
		username = os.Getenv("USERNAME")
		password = os.Getenv("PASSWORD")
		keys = append(keys, &username, &password)
	}

	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	mailFromAddress = os.Getenv("MAIL_FROM")

	for _, s := range keys {
		if *s == "" {
			return responses.ErrEmptyValue
		}
	}
	smtpUrl = fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	return
}

// sendMailSmtps sends an email using smtp over TLS, with smtp authentication.
func sendMailSmtps(mailToAddress string, message []byte) error {
	auth := smtp.PlainAuth("", username, password, smtpHost)
	return smtp.SendMail(smtpUrl, auth, mailFromAddress, []string{mailToAddress}, message)
}

// sendMailSmtp sends an email using smtp without authentication.
// This should only be used when the smtp server and the itpg-backend
// binary are running on the same machine.
func sendMailSmtp(mailToAddress string, message []byte) error {
	c, err := smtp.Dial(smtpUrl)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = c.Mail(mailFromAddress); err != nil {
		return err
	}

	if err = c.Rcpt(mailToAddress); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

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

// makeConfCodeMessage creates the confirmation code email to be sent.
func makeConfCodeMessage(mailToAddress, confirmationCode string) []byte {
	return []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Confirmation Code\r\n\r\nHello %s,\r\n\nYour confirmation code: %s\r\n\nUse this code to complete your registration on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, mailFromAddress, time.Now().Format(time.RFC1123Z), mailToAddress, confirmationCode))
}

// makeResetCodeMessage creates the reset password email to be sent.
func makeResetCodeMessage(mailToAddress, resetLink string) []byte {
	return []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Password Reset Code\r\n\r\nHello %s,\r\n\nYour password reset link: %s\r\n\nUse this code to reset your password on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, mailFromAddress, time.Now().Format(time.RFC1123Z), mailToAddress, resetLink))
}
