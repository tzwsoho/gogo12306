package captcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type CaptchaResult struct {
	Msg    string `json:"msg"`
	Result []int  `json:"result,omitempty"`
}

func NeedCaptcha(jar *cookiejar.Jar) (isNeed bool, err error) {
	const (
		url0    = "https://%s/otn/login/conf"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	req, _ := http.NewRequest("GET", fmt.Sprintf(url0, cdn.GetCDN()), nil)
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
		StudentControl  int    `json:"stu_control"`
		OtherControl    int    `json:"other_control"`
	}

	type NeedCaptchaInfo struct {
		NeedCaptchaData `json:"data"`
	}

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

	// 预售天数
	config.Cfg.StudentPresellDays = info.StudentControl
	config.Cfg.OtherPresellDays = info.OtherControl

	return isNeed, nil
}

func GetCaptcha(jar *cookiejar.Jar) (res string, err error) {
	const (
		url     = "https://%s/passport/captcha/captcha-image64?login_site=E&module=login&rand=sjrand&_=%f"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN(), rand.Float32()), nil)
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

func GetCaptchaResult(jar *cookiejar.Jar, base64Img string, result *CaptchaResult) (err error) {
	esc := url.QueryEscape(base64Img)
	payload := "img=" + esc
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", config.Cfg.OCR.OCRUrl, buf)
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
		1: 坐标 (39, 45)	2: 坐标 (111, 45)	3: 坐标 (183, 45)	4: 坐标 (255, 45)
		5: 坐标 (39, 118)	6: 坐标 (111, 118)	7: 坐标 (183, 118)	8: 坐标 (255, 118)
	*/
	var table [][]int = [][]int{
		{39, 45}, {111, 45}, {183, 45}, {255, 45},
		{39, 118}, {111, 118}, {183, 118}, {255, 118},
	}

	for _, pos := range result.Result {
		ret += strconv.Itoa(table[pos-1][0]+rand.Intn(60)-30) + "," +
			strconv.Itoa(table[pos-1][1]+rand.Intn(60)-30) + ","
	}

	if len(ret) > 1 {
		ret = strings.TrimRight(ret, ",")
	}

	return
}

func VerifyCaptcha(jar *cookiejar.Jar, answer string) (pass bool, err error) {
	const (
		url0    = "https://%s/passport/captcha/captcha-check?answer=%s&rand=sjrand&login_site=E&_=%f"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url0, cdn.GetCDN(), answer, rand.Float64()), nil)
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
		ResultCode    int    `json:"result_code,string"`
		ResultMessage string `json:"result_message"`
	}
	checkResult := CheckResult{}
	if err = json.Unmarshal(body, &checkResult); err != nil {
		logger.Error("解析检查验证码结果失败", zap.ByteString("res", body))

		return
	}

	if checkResult.ResultCode == 4 {
		pass = true
	} else {
		logger.Error("检查验证码结果不通过",
			zap.String("msg", checkResult.ResultMessage),
			zap.Int("code", checkResult.ResultCode),
		)
	}

	return
}
