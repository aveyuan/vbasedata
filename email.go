package vbasedata

import (
	"fmt"
	"log/slog"

	"github.com/wneessen/go-mail"
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
	e *mail.Msg
	c *EmailConfig
	slog *slog.Logger
}

func NewEmail(c *EmailConfig, log *slog.Logger) *Email {
	if log == nil {
		log = slog.Default()
	}
	return &Email{
		e: mail.NewMsg(),
		c: c,
		slog: log,
	}
}

type Msg struct {
	Title    string
	Body     string
	To       string
	BodyType BodyType // 1 text 2 html
}

func (t *Email) SendMsg(msg *Msg) error {
if err := t.e.From(t.c.Form); err != nil {
		t.slog.Error("failed to set From address", "error", err)
		return err
	}
	if err := t.e.To(msg.To); err != nil {
		t.slog.Error("failed to set To address", "error", err)
		return err
	}

	txt := mail.TypeTextPlain
	if msg.BodyType == HtmlBodyType {
		txt = mail.TypeTextHTML
	}
	t.e.Subject("This is my first mail with go-mail!")
	t.e.SetBodyString(txt, msg.Body)
	client, err := mail.NewClient(fmt.Sprintf("%v:%v", t.c.Host, t.c.Port), mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(t.c.Username), mail.WithPassword(t.c.Password))
	if err != nil {
		t.slog.Error("failed to create mail client", "error", err)
		return err
	}
	if err := client.DialAndSend(t.e); err != nil {
		t.slog.Error("failed to send mail", "error", err)
		return err
	}

	return  nil
}
