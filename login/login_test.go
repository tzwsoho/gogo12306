package login_test

import (
	"gogo12306/captcha"
	"gogo12306/cookie"
	"gogo12306/logger"
	"gogo12306/login"
	"net/http/cookiejar"
	"testing"
)

func TestLogin(t *testing.T) {
	const (
		/*
			方法 1 - 使用 SELENIUM 获取，需要预先安装 Chrome 浏览器，需要下载 ChromeDriver，大版本需和 Chrome 一致（http://chromedriver.storage.googleapis.com/index.html）
			方法 2 - 使用 https://kyfw.12306.cn/otn/HttpZF/logdevice 获取，不太稳定，若一直报错请使用其他方法获取
			方法 3 - 打开浏览器按 F12 打开开发者工具并切换到 “网络” 标签页，打开 https://www.12306.cn，找到任意请求的请求标头里面 “Cookie” 下的 RAIL_EXPIRATION 和 RAIL_DEVICEID 的值
		*/
		GetCookieMethod int = 1

		ChromeDriverPath string = "../chromedriver"
		RailExpiration   string = ""
		RailDeviceID     string = ""

		// 自建 12306 验证码 OCR 网址，自建方法参考: https://py12306-helper.pjialin.com/
		OCRURL = ""

		USERNAME = ""
		PASSWORD = ""
	)

	logger.Init(true, "test.log", "info", 1024, 7)

	var (
		err error
		jar *cookiejar.Jar
	)
	if jar, err = cookiejar.New(nil); err != nil {
		t.Error(err.Error())
		return
	}

	if err = cookie.SetCookie(jar, GetCookieMethod, ChromeDriverPath, RailExpiration, RailDeviceID); err != nil {
		t.Error(err.Error())
		return
	}

	var (
		base64Img     string
		captchaResult captcha.CaptchaResult
		pass          bool
	)
	// 获取验证码图像
	if base64Img, err = captcha.GetCaptcha(jar); err != nil {
		t.Error(err.Error())
		return
	}

	// 自动识别验证码并获取结果
	if err = captcha.GetCaptchaResult(jar, OCRURL, base64Img, &captchaResult); err != nil {
		t.Error(err.Error())
		return
	}

	// 将结果转化为坐标点
	answer := captcha.ConvertCaptchaResult(&captchaResult)

	// 校验验证码
	if pass, err = captcha.VerifyCaptcha(jar, answer); err != nil || !pass {
		t.Error(err.Error())
		return
	}

	// 登录
	if err = login.DoLogin(jar, USERNAME, PASSWORD, answer); err != nil {
		t.Error(err.Error())
		return
	}

	// 授权并获取用户信息
	var newapptk string
	if newapptk, err = login.Auth(jar); err != nil {
		t.Error(err.Error())
		return
	}

	if err = login.GetUserInfo(jar, newapptk); err != nil {
		t.Error(err.Error())
		return
	}

	// 获取乘客列表
	if err = login.GetPassengerList(jar); err != nil {
		t.Error(err.Error())
		return
	}
}
