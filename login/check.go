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
	"strings"
	"time"

	"go.uber.org/zap"
)

func CheckLoginStatus(jar *cookiejar.Jar) (logined bool, messages string, err error) {
	const (
		url     = "https://%s/otn/login/checkUser"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	payload := "_json_att="
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url, cdn.GetCDN()), buf)
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

func CheckLoginTimer(jar *cookiejar.Jar) {
	go func() {
		t := time.NewTicker(time.Second)
		for range t.C {
			var (
				logined  bool
				messages string
				err      error
			)
			if logined, messages, err = CheckLoginStatus(jar); err != nil {
				continue
			}

			if !logined {
				logger.Warn("用户已离线，尝试重新登录...", zap.String("错误提示", messages))

				var needCaptcha bool
				if needCaptcha, err = captcha.NeedCaptcha(jar); err != nil {
					continue
				}

				Login(jar, needCaptcha)
			}
		}
	}()
}
