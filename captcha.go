package vbasedata

import (
	"image/color"

	"github.com/aveyuan/base64Captcha"
)

type CaptchaConfig struct {
	Width   int         `json:"width" yaml:"width"`
	Height  int         `json:"height" yaml:"height"`
	Fonts   []string    `json:"fonts" yaml:"fonts"`
	BgColor *color.RGBA `json:"bg_color" yaml:"bg_color"`
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
		c.Fonts = append(c.Fonts, "actionj.ttf")
	}

	if c.BgColor == nil {
		c.BgColor = &color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	driver := (&base64Captcha.DriverMath{
		Width:   c.Width,
		Height:  c.Height,
		Fonts:   c.Fonts,
		BgColor: c.BgColor,
	}).ConvertFonts()

	return &Captcha{
		stor:    stor,
		captcha: base64Captcha.NewCaptcha(driver, stor),
	}
}

// Generate creates a captcha image and stores its answer. The answer is never
// returned to callers, preventing accidental disclosure in an API response.
func (c *Captcha) Generate() (id, image string, err error) {
	id, image, _, err = c.captcha.Generate()
	return id, image, err
}

// Verify validates and consumes a captcha answer.
func (c *Captcha) Verify(id, value string) bool {
	return c.stor.Verify(id, value, true)
}
