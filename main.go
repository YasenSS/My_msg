package main

import (
	"net/smtp"

	"github.com/jordan-wright/email"
)

// tzmstwrrisoziceb
const (
	emailAcount   = "yasenoys@qq.com"
	emailAuthCode = "tzmstwrrisoziceb"
	emailSvr      = "smtp.qq.com:587"
	emailHost     = "smtp.qq.com"
	emailReceiver = "244219338@qq.com"
)

func SendEmail(to string, subject string, text string) {
	em := email.NewEmail()
	em.From = emailAcount
	em.To = []string{to}
	em.Subject = subject
	em.Text = []byte(text)

	em.Send(emailSvr, smtp.PlainAuth("", emailAcount, emailAuthCode, emailHost))
}

func main() {
	SendEmail(emailReceiver, "It's a test", "hello 244219338")
}
