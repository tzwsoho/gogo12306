package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"go.uber.org/zap"
)

func GetUserInfo(jar *cookiejar.Jar, newapptk string) (err error) {
	const (
		url0    = "https://%s/otn/uamauthclient"
		referer = "https://kyfw.12306.cn/otn/passport?redirect=/otn/login/userLogin"
	)
	tk := url.QueryEscape(newapptk)

	payload := "tk=" + tk
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("获取用户信息请求错误", zap.Error(err))

		return err
	} else if !ok {
		logger.Error("获取用户信息请求失败", zap.ByteString("res", body))

		return errors.New("get user info failure")
	}

	type UserInfo struct {
		ResultCode    int    `json:"result_code"`
		ResultMessage string `json:"result_message"`
		AppTK         string `json:"apptk"`
		Username      string `json:"username,omitempty"`
	}

	result := UserInfo{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析获取用户信息错误", zap.ByteString("res", body), zap.Error(err))

		return err
	}

	if result.ResultCode == 0 {
		logger.Debug("用户登录成功", zap.String("用户名", result.Username), zap.ByteString("body", body))
	} else {
		logger.Error("用户登录失败", zap.Int("code", result.ResultCode), zap.String("msg", result.ResultMessage))
	}

	return
}

func GetPassengerList(jar *cookiejar.Jar) (err error) {
	return
}
