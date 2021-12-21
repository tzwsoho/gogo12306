package login

import (
	"bytes"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"

	"go.uber.org/zap"
)

func Auth(jar *cookiejar.Jar) (err error) {
	const (
		url1     = "https://%s/otn/resources/login.html"
		referer1 = "https://kyfw.12306.cn/otn/view/index.html"
	)

	req1, _ := http.NewRequest("GET", fmt.Sprintf(url1, cdn.GetCDN()), nil)
	req1.Header.Set("Referer", referer1)
	httpcli.DefaultHeaders(req1)

	var (
		body1 []byte
		ok1   bool
	)
	if body1, ok1, _, err = httpcli.DoHttp(req1, jar); err != nil {
		logger.Error("授权登录错误", zap.Error(err))

		return err
	} else if !ok1 {
		logger.Error("授权登录失败", zap.ByteString("res", body1))

		return errors.New("login auth failure")
	}

	const (
		url2     = "https://%s/passport/web/auth/uamtk-static"
		referer2 = "https://kyfw.12306.cn/otn/resources/login.html"
	)

	payload := "appid=otn"
	buf := bytes.NewBuffer([]byte(payload))

	req2, _ := http.NewRequest("POST", fmt.Sprintf(url2, cdn.GetCDN()), buf)
	req2.Header.Set("Referer", referer2)
	httpcli.DefaultHeaders(req2)

	var (
		body2 []byte
		ok2   bool
	)
	if body2, ok2, _, err = httpcli.DoHttp(req2, jar); err != nil {
		logger.Error("授权错误", zap.Error(err))

		return err
	} else if !ok2 {
		logger.Error("授权失败", zap.ByteString("res", body2))

		return errors.New("auth failure")
	}

	return
}
