package login

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/captcha"
	"gogo12306/cdn"
	"gogo12306/common"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/tjfoc/gmsm/sm4"
	"go.uber.org/zap"
)

func DoLogin(jar *cookiejar.Jar, username, password, answer string) (err error) {
	// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js 关键词: popup_loginForUam 函数

	const (
		url0    = "https://%s/passport/web/login"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
		sm4key  = "tiekeyuankp12306"
	)
	var encPwd []byte
	if encPwd, err = sm4.Sm4Ecb([]byte(sm4key), []byte(password), true); err != nil {
		logger.Error("SM4 加密密码失败", zap.Error(err))

		return
	}

	payload := url.Values{}
	payload.Add("username", username)
	payload.Add("password", "@"+base64.StdEncoding.EncodeToString(encPwd))
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

	logger.Debug("登录成功!!!", zap.String("uamtk", result.UAMTK))
	return
}

func DoLoginWithoutCaptcha(jar *cookiejar.Jar, username, password string) (err error) {
	const (
		url0    = "https://%s/otn/login/loginAysnSuggest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
		sm4key  = "tiekeyuankp12306"
	)
	var encPwd []byte
	if encPwd, err = sm4.Sm4Ecb([]byte(sm4key), []byte(password), true); err != nil {
		logger.Error("SM4 加密密码失败", zap.Error(err))

		return
	}

	payload := url.Values{}
	payload.Add("loginUserDTO.user_name", username)
	payload.Add("userDTO.password", "@"+base64.StdEncoding.EncodeToString(encPwd))

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

func Login(jar *cookiejar.Jar) (err error) {
	common.CheckOperationPeriod()

	var needCaptcha bool
	if needCaptcha, err = captcha.NeedCaptcha(jar); err != nil {
		return
	}

	// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js 关键词: popup_login 函数
	if needCaptcha { // 需要验证码登录
		var (
			base64Img string
			pass      bool
		)
		// 获取验证码图像
		if base64Img, err = captcha.GetCaptcha(jar); err != nil {
			return
		}

		// 自动识别验证码并获取结果
		var (
			result []int
			answer string
		)
		if _, answer, err = captcha.GetCaptchaResult(jar, config.Cfg.Login.OCRUrl, base64Img); err != nil {
			return
		}

		if answer == "" {
			logger.Error("ConvertCaptchaResult 转换坐标失败", zap.Ints("result", result))
			return
		}

		// 校验验证码
		if pass, err = captcha.VerifyCaptcha(jar, answer); err != nil || !pass {
			return
		}

		// 登录
		if err = DoLogin(jar, config.Cfg.Login.Username, config.Cfg.Login.Password, answer); err != nil {
			return
		}

		// 授权
		var tk string
		if tk, err = Auth(jar); err != nil {
			return
		}

		// 获取用户信息
		if err = GetUserInfo(jar, tk); err != nil {
			return
		}
	} else { // 无需验证码登录
		if err = DoLoginWithoutCaptcha(jar, config.Cfg.Login.Username, config.Cfg.Login.Password); err != nil {
			return
		}
	}

	// 获取乘客列表
	time.Sleep(time.Second)
	if err = GetPassengerList(jar); err != nil {
		return
	}

	return
}
