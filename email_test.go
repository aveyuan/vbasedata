package vbasedata

import (
	"testing"
)

func TestMail(t *testing.T) {
	e := NewEmail(&EmailConfig{
		Username: "0000", //有些邮箱需要添加账号
		Password: "",     //密码
		Host:     "smtp.qq.com",
		Port:     "465",
		Form:     "test<0000@qq.com>",
		Tls:      true,
	})
	if err := e.SendMsg(&Msg{
		Title:    "test",
		Body:     "this is test",
		To:       "0000@qq.com",
		BodyType: TextBodyType,
	}); err != nil {
		t.Fatal(err)
	}
}
