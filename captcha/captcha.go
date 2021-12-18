package captcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"math/rand"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

func GetCaptcha() (res string, err error) {
	const url = "https://kyfw.12306.cn/passport/captcha/captcha-image64?login_site=E&module=login&rand=sjrand&_=%f"
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, rand.Float32()), nil)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req); err != nil {
		logger.Error("获取验证码错误", zap.Error(err))

		return "", err
	} else if !ok {
		logger.Error("获取验证码失败", zap.ByteString("res", body))

		return "", errors.New("get captcha failure")
	}

	type captcha struct {
		Image string `json:"image"`
	}
	cap := captcha{}
	if err = json.Unmarshal(body, &cap); err != nil {
		logger.Error("解析验证码失败", zap.ByteString("res", body))

		return "", errors.New("analyze captcha failure")
	}

	return cap.Image, nil
}

func GetCaptchaResult(ocrUrl, base64Img string) (res string, err error) {
	esc := url.QueryEscape(base64Img)
	payload := "img=" + esc
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", ocrUrl, buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req); err != nil {
		logger.Error("获取验证码结果错误", zap.Error(err))

		return "", err
	} else if !ok {
		logger.Error("获取验证码结果失败", zap.ByteString("res", body))

		return "", errors.New("get captcha result failure")
	}

	return string(body), nil
}
