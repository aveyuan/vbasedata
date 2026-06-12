package vbasedata

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewCaptcha_Defaults(t *testing.T) {
	c := &CaptchaConfig{}
	NewCaptcha(c, NewLruCache(10, time.Minute))

	if c.Width != 320 {
		t.Errorf("Width default = %d, want 320", c.Width)
	}
	if c.Height != 120 {
		t.Errorf("Height default = %d, want 120", c.Height)
	}
	if len(c.Fonts) == 0 || c.Fonts[0] != "actionj.ttf" {
		t.Errorf("Fonts default = %v, want [actionj.ttf]", c.Fonts)
	}
	if c.BgColor == nil || c.BgColor.A != 255 {
		t.Errorf("BgColor = %+v, want opaque (A=255)", c.BgColor)
	}
}

func TestCaptcha_GenerateAndVerify(t *testing.T) {
	c := NewCaptcha(&CaptchaConfig{}, NewLruCache(10, time.Minute))
	ctx := context.Background()

	id, b64s, answer, err := c.GetCaptCha(ctx)
	if err != nil {
		t.Fatalf("GetCaptCha: %v", err)
	}
	if id == "" || answer == "" {
		t.Fatalf("empty id (%q) or answer (%q)", id, answer)
	}
	if !strings.HasPrefix(b64s, "data:image") {
		t.Errorf("b64s does not look like a data URI: %.20q", b64s)
	}

	// Wrong answer fails.
	if c.Verify(ctx, id, answer+"_wrong") {
		t.Error("Verify should fail for a wrong answer")
	}
}

func TestCaptcha_VerifyConsumes(t *testing.T) {
	c := NewCaptcha(&CaptchaConfig{}, NewLruCache(10, time.Minute))
	ctx := context.Background()

	id, _, answer, err := c.GetCaptCha(ctx)
	if err != nil {
		t.Fatalf("GetCaptCha: %v", err)
	}

	if !c.Verify(ctx, id, answer) {
		t.Fatal("Verify should succeed for the correct answer")
	}
	// The store clears on verify, so a second attempt must fail.
	if c.Verify(ctx, id, answer) {
		t.Error("second Verify should fail (answer already consumed)")
	}
}
