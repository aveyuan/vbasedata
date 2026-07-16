package vbasedata

import (
	"os"
	"strings"
	"testing"
)

// TestEmail_SendMsg 通过真实 SMTP 发送一封测试邮件。
// 优先使用 VB_TEST_SMTP_* 环境变量；未设置时回退到默认值，
// 若默认设置发送失败则跳过（视为本地无可用 SMTP），显式配置了才判定失败。
//
// 可用环境变量：
//
//	VB_TEST_SMTP_HOST  SMTP 主机（默认 127.0.0.1）
//	VB_TEST_SMTP_PORT  端口（默认 25）
//	VB_TEST_SMTP_FROM  发件人，对应 EmailConfig.Form（默认 test@example.com）
//	VB_TEST_SMTP_TO    收件人（默认 test@example.com）
//	VB_TEST_SMTP_USER  认证用户名（可选）
//	VB_TEST_SMTP_PASS  认证密码（可选）
//	VB_TEST_SMTP_TLS   置为 1/true 时启用隐式 SSL/TLS（可选）
func TestEmail_SendMsg(t *testing.T) {
	requireIntegration(t)
	host, d1 := testEnvDefault("VB_TEST_SMTP_HOST", "127.0.0.1")
	port, d2 := testEnvDefault("VB_TEST_SMTP_PORT", "25")
	from, d3 := testEnvDefault("VB_TEST_SMTP_FROM", "test@example.com")
	to, d4 := testEnvDefault("VB_TEST_SMTP_TO", "test@example.com")
	user := os.Getenv("VB_TEST_SMTP_USER")
	pass := os.Getenv("VB_TEST_SMTP_PASS")
	tlsEnv := strings.TrimSpace(os.Getenv("VB_TEST_SMTP_TLS"))
	tls := tlsEnv == "1" || strings.EqualFold(tlsEnv, "true")
	usingDefault := d1 || d2 || d3 || d4

	e := NewEmail(&EmailConfig{
		Username: user,
		Password: pass,
		Host:     host,
		Port:     port,
		Form:     from,
		Tls:      tls,
	}, discardLogger())

	err := e.SendMsg(&Msg{
		Title:    "vbasedata SMTP test",
		Body:     "this is a test message from vbasedata",
		To:       to,
		BodyType: TextBodyType,
	})
	if err != nil {
		if usingDefault {
			t.Skipf("smtp is not available at %s:%s with default test settings: %v", host, port, err)
		}
		t.Fatalf("SendMsg: %v", err)
	}
}
