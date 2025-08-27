package vbasedata

import (
	"context"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
)

type CaptchaConfig struct {
	Width      int           `json:"width" yaml:"width"`
	Height     int           `json:"height" yaml:"height"`
	Fonts      []string      `json:"fonts" yaml:"fonts"`
	BgColor    *color.RGBA   `json:"bg_color" yaml:"bg_color"`
	StorageLen int           `json:"storage_len" yaml:"storage_len"`
	StroageExp time.Duration `json:"stroage_exp" yaml:"stroage_exp"`
}

type Captcha struct {
	stor    base64Captcha.Store
	captcha *base64Captcha.Captcha
}

func NewCaptcha(c *CaptchaConfig, stor base64Captcha.Store) *Captcha {
	if c.Width == 0 {
		c.Width = 320
	}

	if c.Height == 0 {
		c.Height = 120
	}

	if len(c.Fonts) == 0 {
		c.Fonts = append(c.Fonts, "wqy-microhei.ttc")
	}

	if c.BgColor == nil {
		c.BgColor = &color.RGBA{
			R: 255,
			B: 255,
			G: 255,
		}
	}

	if c.StorageLen == 0 {
		c.StorageLen = 5000
	}

	if c.StroageExp == 0 {
		c.StroageExp = 6 * time.Minute
	}

	dv := &base64Captcha.DriverMath{
		Width:   c.Width,
		Height:  c.Height,
		Fonts:   c.Fonts,
		BgColor: c.BgColor,
	}
	dv = dv.ConvertFonts()

	// 实例化
	return &Captcha{
		stor:    stor,
		captcha: base64Captcha.NewCaptcha(dv, stor),
	}
}

func (r *Captcha) GetCaptCha(ctx context.Context) (id, b64s, answer string, err error) {
	return r.captcha.Generate()
}

func (r *Captcha) Verify(ctx context.Context, id, VerifyValue string) (b bool) {
	return r.stor.Verify(id, VerifyValue, true)
}

