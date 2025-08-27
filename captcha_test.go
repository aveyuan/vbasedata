package vbasedata

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestCaptcha(t *testing.T) {
	c := NewCaptcha(&CaptchaConfig{}, NewLruCache(2, 3*time.Second))
	id, base, ans, err := c.GetCaptCha(context.Background())
	log.Print(id, base, ans, err)
	log.Print(c.Verify(context.Background(), id, ans))

}
