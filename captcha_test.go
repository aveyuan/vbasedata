package vbasedata

import (
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
	store := NewLruCache(10, time.Minute)
	captcha := NewCaptcha(&CaptchaConfig{}, store)

	id, image, err := captcha.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if id == "" {
		t.Fatal("empty captcha id")
	}
	if !strings.HasPrefix(image, "data:image") {
		t.Errorf("image does not look like a data URI: %.20q", image)
	}
	if captcha.Verify(id, "wrong") {
		t.Error("Verify should fail for a wrong answer")
	}
}

func TestCaptcha_VerifyConsumes(t *testing.T) {
	store := NewLruCache(10, time.Minute)
	captcha := NewCaptcha(&CaptchaConfig{}, store)

	id, _, err := captcha.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	answer := store.Get(id, false)
	if answer == "" {
		t.Fatal("captcha answer was not stored")
	}
	if !captcha.Verify(id, answer) {
		t.Fatal("Verify should succeed for the correct answer")
	}
	if captcha.Verify(id, answer) {
		t.Error("second Verify should fail (answer already consumed)")
	}
}
