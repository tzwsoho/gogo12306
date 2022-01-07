package captcha_test

import (
	"gogo12306/captcha"
	"gogo12306/logger"
	"net/http/cookiejar"
	"testing"

	"go.uber.org/zap"
)

func TestCaptcha(t *testing.T) {
	// 自建 12306 验证码 OCR 网址，自建方法参考: https://py12306-helper.pjialin.com/
	const OCRURL = ""

	logger.Init(true, "test.log", "info", 1024, 7)

	var (
		err       error
		jar       *cookiejar.Jar
		base64Img string
		pass      bool
	)
	if jar, err = cookiejar.New(nil); err != nil {
		t.Error(err.Error())
		return
	}

	// 获取校验码图片 BASE64
	if base64Img, err = captcha.GetCaptcha(jar); err != nil {
		t.Error(err.Error())
		return
	}

	// 自动识别校验码
	var (
		result []int
		answer string
	)
	if result, answer, err = captcha.GetCaptchaResult(jar, OCRURL, base64Img); err != nil {
		t.Error(err.Error())
		return
	}

	// 验证校验码结果
	if pass, err = captcha.VerifyCaptcha(jar, answer); err != nil {
		t.Error(err.Error())
		return
	}

	logger.Info("校验码验证结果",
		zap.String("校验码图片信息", base64Img),
		zap.Any("校验码 OCR 结果", result),
		zap.String("转化后坐标点", answer),
		zap.Bool("校验码验证是否通过", pass),
	)
}
