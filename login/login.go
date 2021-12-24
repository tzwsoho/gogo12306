package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/captcha"
	"gogo12306/cdn"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"go.uber.org/zap"
)

func DoLogin(jar *cookiejar.Jar, answer string) (err error) {
	const (
		url0    = "https://%s/passport/web/login"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	username := url.QueryEscape(config.Cfg.Login.Username)
	password := url.QueryEscape(config.Cfg.Login.Password)
	ans := url.QueryEscape(answer)

	payload := "username=" + username +
		"&password=" + password +
		"&appid=otn" +
		"&answer" + ans
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("登录请求错误", zap.Error(err))

		return err
	} else if !ok {
		logger.Error("登录请求失败", zap.ByteString("res", body))

		return errors.New("login failure")
	}

	type LoginResult struct {
		ResultCode    int    `json:"result_code"`
		ResultMessage string `json:"result_message"`
		UAMTK         string `json:"uamtk"`
	}

	result := LoginResult{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析登录返回信息错误", zap.ByteString("res", body), zap.Error(err))

		return err
	}

	if result.ResultCode != 0 {
		logger.Error("登录失败", zap.String("错误信息", result.ResultMessage))

		return errors.New("login failure")
	}

	// logger.Debug("登录成功！", zap.String("uamtk", result.UAMTK))
	return
}

func DoLoginWithoutCaptcha(jar *cookiejar.Jar) (err error) {
	const (
		url0    = "https://%s/otn/login/loginAysnSuggest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	username := url.QueryEscape(config.Cfg.Login.Username)
	password := url.QueryEscape(config.Cfg.Login.Password)

	payload := "loginUserDTO.user_name=" + username +
		"&userDTO.password=" + password
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("无验证码登录请求错误", zap.Error(err))

		return err
	} else if !ok {
		logger.Error("无验证码登录请求失败", zap.ByteString("res", body))

		return errors.New("login without captcha failure")
	}

	type LoginWithoutCaptchaData struct {
		LoginCheck string `json:"loginCheck"`
	}

	type LoginWithoutCaptchaResult struct {
		Data LoginWithoutCaptchaData `json:"data"`
	}

	result := LoginWithoutCaptchaResult{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析无验证码登录返回信息错误", zap.ByteString("res", body), zap.Error(err))

		return err
	}

	if result.Data.LoginCheck != "Y" {
		logger.Error("无验证码登录失败", zap.ByteString("res", body))

		return errors.New("login failure")
	}

	// logger.Debug("无验证码登录成功！")
	return
}

func Login(jar *cookiejar.Jar, needCaptcha bool) (err error) {
	if err = SetCookie(jar); err != nil {
		return
	}

	if needCaptcha { // 需要验证码登录
		var (
			base64Img     string
			captchaResult captcha.CaptchaResult
			pass          bool
		)
		if base64Img, err = captcha.GetCaptcha(jar); err != nil {
			return
		}

		if err = captcha.GetCaptchaResult(jar, base64Img, &captchaResult); err != nil {
			return
		}

		answer := captcha.ConvertCaptchaResult(&captchaResult)

		if pass, err = captcha.VerifyCaptcha(jar, answer); err != nil || !pass {
			return
		}

		if err = DoLogin(jar, answer); err != nil {
			return
		}

		var newapptk string
		if newapptk, err = Auth(jar); err != nil {
			return
		}

		if err = GetUserInfo(jar, newapptk); err != nil {
			return
		}
	} else { // 无需验证码登录
		if err = DoLoginWithoutCaptcha(jar); err != nil {
			return
		}
	}

	// if err = GetPassengerList(jar); err != nil {
	// 	return
	// }

	return
}
