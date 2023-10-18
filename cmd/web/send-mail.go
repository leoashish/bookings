package main

import (
	"fmt"
	"github.com/leoashish99/bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"log"
	"os"
	"strings"
	"time"
)

func ListenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			SendMessage(msg)
		}
	}()
}

func SendMessage(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		errorLog.Println(err)
	}
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)

	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := os.ReadFile(fmt.Sprintf("../../email-templates/%s", m.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}
		mailTemplate := string(data)
		messageToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, messageToSend)
	}

	err = email.Send(client)

	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent!!")
	}
}
