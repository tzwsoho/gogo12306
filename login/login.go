package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"

	"go.uber.org/zap"
)

func NeedCaptcha(jar *cookiejar.Jar) (isNeed bool, err error) {
	const (
		url     = "https://%s/otn/login/conf"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN()), nil)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("获取登录是否需要验证码错误", zap.Error(err))

		return false, err
	} else if !ok {
		logger.Error("获取登录是否需要验证码失败", zap.ByteString("res", body))

		return false, errors.New("get need captcha failure")
	}

	type NeedCaptchaData struct {
		IsLoginPassCode string `json:"is_login_passCode"`
	}

	type NeedCaptchaInfo struct {
		NeedCaptchaData `json:"data"`
	}

	logger.Debug("Is Need Captcha Body", zap.ByteString("body", body))

	info := NeedCaptchaInfo{}
	if err = json.Unmarshal(body, &info); err != nil {
		logger.Error("解析登录是否需要验证码信息错误", zap.ByteString("res", body), zap.Error(err))

		return false, err
	}

	if info.IsLoginPassCode == "N" {
		isNeed = false
	} else {
		isNeed = true
	}

	return isNeed, nil
}

func Login() (err error) {
	var jar *cookiejar.Jar
	if jar, err = cookiejar.New(nil); err != nil {
		logger.Error("创建 Jar 错误", zap.Error(err))
		return
	}

	if err = SetCookie(jar); err != nil {
		return
	}

	var isNeed bool
	if isNeed, err = NeedCaptcha(jar); err != nil {
		return
	}

	if isNeed { // 需要验证码登录
		logger.Debug("Need")

	} else { // 无需验证码登录
		logger.Debug("No need")

	}

	return
}
