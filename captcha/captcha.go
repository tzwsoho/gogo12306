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
	"net/http/cookiejar"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

type CaptchaResult struct {
	Msg    string `json:"msg"`
	Result []int  `json:"result,omitempty"`
}

func GetCaptcha(jar *cookiejar.Jar) (res string, err error) {
	const (
		url     = "https://kyfw.12306.cn/passport/captcha/captcha-image64?login_site=E&module=login&rand=sjrand&_=%f"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, rand.Float32()), nil)
	req.Header.Add("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
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

func GetCaptchaResult(jar *cookiejar.Jar, ocrUrl, base64Img string, result *CaptchaResult) (err error) {
	esc := url.QueryEscape(base64Img)
	payload := "img=" + esc
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", ocrUrl, buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("获取验证码结果错误", zap.Error(err))

		return err
	} else if !ok {
		logger.Error("获取验证码结果失败", zap.ByteString("res", body))

		return errors.New("get captcha result failure")
	}

	if err = json.Unmarshal(body, result); err != nil {
		logger.Error("解析验证码结果错误", zap.ByteString("res", body), zap.Error(err))

		return err
	}

	return
}

func ConvertCaptchaResult(result *CaptchaResult) (ret string) {
	/*
		验证码选项图像编号：
		*****************
		| 1 | 2 | 3 | 4 |
		*****************
		| 5 | 6 | 7 | 8 |
		*****************

		按序号取中点坐标：
		1: 坐标 (39, 75)	2: 坐标 (111, 75)	3: 坐标 (183, 75)	4: 坐标 (255, 75)
		5: 坐标 (39, 148)	6: 坐标 (111, 148)	7: 坐标 (183, 148)	8: 坐标 (255, 148)
	*/
	var table [][]string = [][]string{
		{"39", "75"}, {"111", "75"}, {"183", "75"}, {"255", "75"},
		{"39", "148"}, {"111", "148"}, {"183", "148"}, {"255", "148"},
	}

	for _, pos := range result.Result {
		ret += table[pos-1][0] + "," + table[pos-1][1] + ","
	}

	if len(ret) > 1 {
		ret = strings.TrimRight(ret, ",")
	}

	return
}

func CheckCaptcha(jar *cookiejar.Jar, answer string) (pass bool, err error) {
	const (
		url     = "https://kyfw.12306.cn/passport/captcha/captcha-check?answer=%s&rand=sjrand&login_site=E&_=%f"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, answer, rand.Float32()), nil)
	req.Header.Add("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("检查验证码错误", zap.Error(err))

		return false, err
	} else if !ok {
		logger.Error("检查验证码失败", zap.ByteString("res", body))

		return false, errors.New("check captcha failure")
	}

	type CheckResult struct {
		ResultCode int `json:"result_code,string"`
	}
	checkResult := CheckResult{}
	if err = json.Unmarshal(body, &checkResult); err != nil {
		logger.Error("解析检查验证码结果失败", zap.ByteString("res", body))

		return
	}

	if checkResult.ResultCode == 4 {
		pass = true
	}

	return
}
