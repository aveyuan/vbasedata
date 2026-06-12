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
	c    *EmailConfig
	slog *slog.Logger
}

func NewEmail(c *EmailConfig, log *slog.Logger) *Email {
	if log == nil {
		log = slog.Default()
	}
	return &Email{
		c:    c,
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
	// 每次发送都新建 Msg，避免复用同一实例导致收件人累加、并发不安全。
	e := mail.NewMsg()
	if err := e.From(t.c.Form); err != nil {
		t.slog.Error("failed to set From address", "error", err)
		return err
	}
	if err := e.To(msg.To); err != nil {
		t.slog.Error("failed to set To address", "error", err)
		return err
	}

	txt := mail.TypeTextPlain
	if msg.BodyType == HtmlBodyType {
		txt = mail.TypeTextHTML
	}
	e.Subject(msg.Title)
	e.SetBodyString(txt, msg.Body)

	opts := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(t.c.Username),
		mail.WithPassword(t.c.Password),
	}
	if t.c.Tls {
		opts = append(opts, mail.WithSSL())
	}
	client, err := mail.NewClient(fmt.Sprintf("%v:%v", t.c.Host, t.c.Port), opts...)
	if err != nil {
		t.slog.Error("failed to create mail client", "error", err)
		return err
	}
	if err := client.DialAndSend(e); err != nil {
		t.slog.Error("failed to send mail", "error", err)
		return err
	}

	return nil
}
