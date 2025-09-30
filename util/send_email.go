package util

import (
	"fmt"
	"net/smtp" // this package handles sending email through SMTP
)

func SendEmail(toEmail, subject, body string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	senderEmail := "mamatasheth@gmail.com"
	senderPassword := "xflvmutvlpksyguc" //app password

	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)) //\r\n\r\n → separates the headers from the body.

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, message)
}
