package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/captcha"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

func CheckLoginStatus(jar *cookiejar.Jar) (logined bool, messages string, err error) {
	const (
		url0    = "https://%s/otn/login/checkUser"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	payload := url.Values{}
	payload.Add("_json_att", "")

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	if body, statusCode, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("检查登录状态错误", zap.Error(err))

		return false, "", err
	} else if statusCode != http.StatusOK {
		logger.Error("检查登录状态失败", zap.Int("statusCode", statusCode), zap.ByteString("res", body))

		return false, "", errors.New("check login status failure")
	}

	type LoginStatusData struct {
		Flag bool `json:"flag"`
	}

	type CheckLoginStatusResult struct {
		Data     LoginStatusData `json:"data"`
		Messages []string        `json:"messages"`
	}

	result := CheckLoginStatusResult{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析检查登录状态信息错误", zap.ByteString("res", body), zap.Error(err))

		return false, "", err
	}

	return result.Data.Flag, strings.Join(result.Messages, ","), nil
}

func CheckAndRelogin(jar *cookiejar.Jar) (err error) {
	var (
		logined  bool
		messages string
	)
	if logined, messages, err = CheckLoginStatus(jar); err != nil {
		return
	}

	if !logined {
		logger.Warn("用户已离线，尝试重新登录...", zap.String("错误提示", messages))

		var needCaptcha bool
		if needCaptcha, err = captcha.NeedCaptcha(jar); err != nil {
			return
		}

		if err = Login(jar, needCaptcha); err != nil {
			return
		}
	}

	return
}

func CheckLoginTimer(jar *cookiejar.Jar) {
	go func() {
		t := time.NewTicker(time.Second * 30) // 检查时间间隔不要太短
		for range t.C {
			CheckAndRelogin(jar)
		}
	}()
}
