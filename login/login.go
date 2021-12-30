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
	"time"

	"go.uber.org/zap"
)

func DoLogin(jar *cookiejar.Jar, answer string) (err error) {
	const (
		url0    = "https://%s/passport/web/login"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	payload := url.Values{}
	payload.Add("username", config.Cfg.Login.Username)
	payload.Add("password", config.Cfg.Login.Password)
	payload.Add("appid", "otn")
	payload.Add("answer", answer)

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	if body, statusCode, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("登录请求错误", zap.Error(err))

		return err
	} else if statusCode != http.StatusOK {
		logger.Error("登录请求失败", zap.Int("statusCode", statusCode), zap.ByteString("res", body))

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

	logger.Debug("登录成功！", zap.String("uamtk", result.UAMTK))
	return
}

func DoLoginWithoutCaptcha(jar *cookiejar.Jar) (err error) {
	const (
		url0    = "https://%s/otn/login/loginAysnSuggest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	payload := url.Values{}
	payload.Add("loginUserDTO.user_name", config.Cfg.Login.Username)
	payload.Add("userDTO.password", config.Cfg.Login.Password)

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	if body, statusCode, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("无验证码登录请求错误", zap.Error(err))

		return err
	} else if statusCode != http.StatusOK {
		logger.Error("无验证码登录请求失败", zap.Int("statusCode", statusCode), zap.ByteString("res", body))

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

	logger.Debug("无验证码登录成功！")
	return
}

func Login(jar *cookiejar.Jar, needCaptcha bool) (err error) {
	now := time.Now()
	if now.Hour() < 6 || now.Hour() >= 23 { // 不在12306 运营期间不能登录或抢票
		return
	}

	if needCaptcha { // 需要验证码登录
		var (
			base64Img     string
			captchaResult captcha.CaptchaResult
			pass          bool
		)
		// 获取验证码图像
		if base64Img, err = captcha.GetCaptcha(jar); err != nil {
			return
		}

		// 自动识别验证码并获取结果
		if err = captcha.GetCaptchaResult(jar, base64Img, &captchaResult); err != nil {
			return
		}

		// 将结果转化为坐标点
		answer := captcha.ConvertCaptchaResult(&captchaResult)

		// 校验验证码
		if pass, err = captcha.VerifyCaptcha(jar, answer); err != nil || !pass {
			return
		}

		// 登录
		if err = DoLogin(jar, answer); err != nil {
			return
		}

		// 授权并获取用户信息
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

	// 获取乘客列表
	if err = GetPassengerList(jar); err != nil {
		return
	}

	return
}
