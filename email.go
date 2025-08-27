package vbasedata

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type EmailConfig struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	Form     string `yaml:"form" json:"form"`
	Tls      bool   `yaml:"tls" json:"tls"`
}

type BodyType string

const (
	TextBodyType = "text"
	HtmlBodyType = "html"
)

type Email struct {
	e *email.Email
	c *EmailConfig
}

func NewEmail(c *EmailConfig) *Email {
	return &Email{
		e: email.NewEmail(),
		c: c,
	}
}

type Msg struct {
	Title    string
	Body     string
	To       string
	BodyType BodyType // 1 text 2 html
}

func (t *Email) SendMsg(msg *Msg) error {
	t.e.From = t.c.Form
	t.e.To = []string{msg.To}
	t.e.Subject = msg.Title
	if msg.BodyType == TextBodyType {
		t.e.Text = []byte(msg.Body)
	}
	if msg.BodyType == HtmlBodyType {
		t.e.HTML = []byte(msg.Body)
	}

	if t.c.Tls {
		return t.e.SendWithTLS(fmt.Sprintf("%v:%v", t.c.Host, t.c.Port), smtp.PlainAuth("", t.c.Username, t.c.Password, t.c.Host), &tls.Config{
			ServerName:         t.c.Host,
			InsecureSkipVerify: false,
		})

	}

	return t.e.Send(fmt.Sprintf("%v:%v", t.c.Host, t.c.Port), smtp.PlainAuth("", t.c.Username, t.c.Password, t.c.Host))

}
