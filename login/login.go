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

// Login 登录
/*
12306 的登录流程非常复杂，主要是使用了主页 https://kyfw.12306.cn/otn/resources/login.html
里面的一个 js 文件 js/login_new_v(xxx).js（xxx 是版本号），
由于本程序需要做到完全自动化登录，
所以在所有 12306 的登录方式中可以做的只有识别图片验证码登录的方式

① 调用 popup_initLogin 函数，主要是页面元素的隐藏和显示，调用 popup_getConf 函数

② 调用 popup_getConf 函数，获取登录配置，并且调用 popup_isLogin 函数：
a. popup_is_uam_login 是否使用统一认证登录界面（输入用户名+密码/扫二维码登录的界面）
b. popup_is_login_passCode 是否使用本地登录（滑动验证/短信校验的界面）
c. popup_is_message_passCode 是否显示短信校验登录方式
d. popup_is_sweep_login 是否显示扫二维码登录入口
e. popup_is_login 是否已登录

③ popup_isLogin 函数有以下逻辑：
if popup_is_uam_login {
	res = http_post('web/auth/uamtk-static')
	if res.result_code == 0 {
		已经登录，重定向页面
		return
	} else if popup_is_sweep_login {
		显示二维码登录入口
	} else {
		隐藏二维码登录入口
	}

	popup_validate()
} else {
	if popup_is_login {
		已经登录，重定向页面
		return
	} else {
		if popup_is_login_passCode {
			if popup_is_message_passCode {
				显示短信校验入口
			} else {
				隐藏短信校验入口
			}
		} else {
			设置登录窗口局中
		}

		popup_validate()
	}
}

④ popup_validate 函数逻辑：
if popup_is_uam_login {
	res = http_post('web/checkLoginVerify')
	if res.login_check_code == 0 {
		res = http_post('web/login', { randCode: 手机校验码 })
		if res.result_code == 0 {
			登录成功，popup_loginCallBack()
			...（后面还有判断是否弹出登录入口等判断）
		} else if res.result_code == 101 {
			需要修改密码才能登陆
		} else if res.result_code in (91, 94, 95, 97) {
			重定向
		} else {
			显示错误信息
		}
	} else ... {
		显示滑动验证/短信校验登录界面，这里还有点不一样：
		如果用户是鼠标点登录按钮，会优先显示滑动验证界面；
		如果用户是用键盘按回车登录，会优先显示短信验证界面。
	}
} else {
	if popup_is_login_passCode {
		选择登陆验证方式
	} else {
		popup_loginForLocation()
	}
}

⑤ popup_loginForLocation 函数：
res = http_post('otn/login/loginAysnSuggest')
if res.loginCheck == 'Y' {
	登录成功，popup_loginCallBack()
}
(后面继续补充...)

*/
func Login(jar *cookiejar.Jar) (err error) {
	common.CheckOperationPeriod()

	var conf *LoginConfResult
	if conf, err = loginConf(jar); err != nil {
		return
	}

	// TODO 以下登录逻辑是参照 py12306 项目而来，需要按 12306 官网源码重新修改逻辑
	if conf.IsLogin { // 用户已登录
		return
	}

	if conf.IsUAMLogin { // 需要验证码登录
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
	if err = GetPassengerList(jar); err != nil {
		return
	}

	return
}
