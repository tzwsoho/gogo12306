package cookie_test

import (
	"gogo12306/cookie"
	"gogo12306/logger"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"go.uber.org/zap"
)

func TestSetCookie(t *testing.T) {
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
	)

	logger.Init(true, "test.log", "info", 1024, 7)

	var (
		err error
		jar *cookiejar.Jar
		u   *url.URL
	)
	if jar, err = cookiejar.New(nil); err != nil {
		t.Error(err.Error())
		return
	}

	if err = cookie.SetCookie(jar, GetCookieMethod, ChromeDriverPath, RailExpiration, RailDeviceID); err != nil {
		t.Error(err.Error())
		return
	}

	if u, err = url.Parse("https://kyfw.12306.cn/"); err != nil {
		t.Error(err.Error())
		return
	}

	logger.Info("当前 Cookie", zap.Any("cookies", jar.Cookies(u)))
}
