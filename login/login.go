package login

import (
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"

	"go.uber.org/zap"
)

func NeedCaptcha() (isNeed bool, err error) {
	const (
		url     = "https://%s/otn/login/conf"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	req, _ := http.NewRequest("POST", fmt.Sprintf(url, cdn.GetCDN()), nil)
	req.Header.Add("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req); err != nil {
		logger.Error("获取登录是否需要验证码错误", zap.Error(err))

		return false, err
	} else if !ok {
		logger.Error("获取登录是否需要验证码失败", zap.ByteString("res", body))

		return false, errors.New("get need captcha failure")
	}

	logger.Debug("Need Captcha", zap.ByteString("body", body))

	return true, nil
}

func Login() (err error) {
	NeedCaptcha()
	return
}
