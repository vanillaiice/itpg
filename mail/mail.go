package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type SmtpClient struct {
	host     string
	url      string
	mailFrom string
	username string
	password string
	secure   bool
}

func NewClient(envPath string, secure bool) (*SmtpClient, error) {
	godotenv.Load(envPath) //nolint:errcheck

	var client SmtpClient

	keys := []string{"MAIL_FROM", "SMTP_HOST", "SMTP_PORT"}

	keysMap := map[string]string{
		keys[0]: os.Getenv(keys[0]),
		keys[1]: os.Getenv(keys[1]),
		keys[2]: os.Getenv(keys[2]),
	}

	if secure {
		client.secure = secure

		keys = append(keys, "USERNAME", "PASSWORD")

		keysMap["USERNAME"] = os.Getenv("USERNAME")
		keysMap["PASSWORD"] = os.Getenv("PASSWORD")
	}

	for _, k := range keys {
		if _, ok := keysMap[k]; !ok {
			return nil, fmt.Errorf("missing %s", k)
		}
	}

	client.host = keysMap["SMTP_HOST"]
	client.url = fmt.Sprintf("%s:%s", client.host, keysMap["SMTP_PORT"])
	client.mailFrom = keysMap["MAIL_FROM"]
	client.username = keysMap["USERNAME"]
	client.password = keysMap["PASSWORD"]

	return &client, nil
}

func (c *SmtpClient) SendMail(mailToAddress string, message []byte) error {
	if c.secure {
		return sendMailSmtps(c.username, c.password, c.host, c.url, c.mailFrom, mailToAddress, message)
	} else {
		return sendMailSmtp(c.url, c.mailFrom, mailToAddress, message)
	}
}

// sendMailSmtps sends an email using smtp over TLS, with smtp authentication.
func sendMailSmtps(username, password, host, smtpUrl, mailFromAddress, mailToAddress string, message []byte) error {
	auth := smtp.PlainAuth("", username, password, host)
	return smtp.SendMail(smtpUrl, auth, mailFromAddress, []string{mailToAddress}, message)
}

// sendMailSmtp sends an email using smtp without authentication.
// This should only be used when the smtp server and the itpg-backend
// binary are running on the same machine.
func sendMailSmtp(url, mailFromAddress, mailToAddress string, message []byte) error {
	c, err := smtp.Dial(url)
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

// MakeConfCodeMessage creates the confirmation code email to be sent.
func (c *SmtpClient) MakeConfCodeMessage(mailToAddress, confirmationCode string) []byte {
	return []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Confirmation Code\r\n\r\nHello %s,\r\n\nYour confirmation code: %s\r\n\nUse this code to complete your registration on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, c.mailFrom, time.Now().Format(time.RFC1123Z), mailToAddress, confirmationCode))
}

// MakeResetCodeMessage creates the reset password email to be sent.
func (c *SmtpClient) MakeResetCodeMessage(mailToAddress, resetLink string) []byte {
	return []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nDate: %s\r\nSubject: ITPG Account Password Reset Code\r\n\r\nHello %s,\r\n\nYour password reset link: %s\r\n\nUse this code to reset your password on itpg.cc.\r\n\nThanks,\r\nITPG Team\r\n\r\nThis is an auto-generated email. Please do not reply to it.\r\n", mailToAddress, c.mailFrom, time.Now().Format(time.RFC1123Z), mailToAddress, resetLink))
}
