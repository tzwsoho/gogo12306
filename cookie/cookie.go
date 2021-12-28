package cookie

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"go.uber.org/zap"
)

func GetBySelenium(jar *cookiejar.Jar) (railExpiration, railDeviceID string, err error) {
	opts := []selenium.ServiceOption{}
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	// 禁止加载图片，加快渲染速度
	imagCaps := map[string]interface{}{
		"profile.managed_default_content_settings.images": 2,
	}

	chromeCaps := chrome.Capabilities{
		Prefs: imagCaps,
		Path:  "",
		Args: []string{
			"--headless",
			"--no-sandbox",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36 Edg/96.0.1054.62",
		},
	}

	caps.AddChrome(chromeCaps)

	const (
		port = 4444
		url  = "https://www.12306.cn/index/index.html"
	)

	service, err := selenium.NewChromeDriverService(config.Cfg.Login.ChromeDriverPath, port, opts...)
	if err != nil {
		logger.Error("启动 ChromeDriver 失败，请检查 ChromeDriver 是否已正确安装", zap.Error(err))
		return
	}
	defer service.Stop()

	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		logger.Error("启动 Chrome 浏览器失败，请检查 Chrome 浏览器是否已正确安装", zap.Error(err))
		return
	}
	defer webDriver.Quit()

	if err = webDriver.Get(url); err != nil {
		logger.Error("导航到 12306 官网主页失败", zap.Error(err))
		return
	}

	var cookieExpiration, cookieDeviceID selenium.Cookie
	if cookieExpiration, err = webDriver.GetCookie("RAIL_EXPIRATION"); err != nil {
		logger.Error("获取 RAIL_EXPIRATION 失败", zap.Error(err))
		return
	}

	if cookieDeviceID, err = webDriver.GetCookie("RAIL_DEVICEID"); err != nil {
		logger.Error("获取 RAIL_DEVICEID 失败", zap.Error(err))
		return
	}

	return cookieExpiration.Value, cookieDeviceID.Value, nil
}

func GetHttpZF(jar *cookiejar.Jar) (railExpiration, railDeviceID string, err error) {
	const (
		url = "https://kyfw.12306.cn/otn/HttpZF/logdevice"
	)
	req, _ := http.NewRequest("GET", url, nil)
	httpcli.DefaultHeaders(req)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("访问获取浏览器 ID 网站错误", zap.Error(err))

		return
	} else if !ok {
		logger.Error("访问获取浏览器 ID 网站失败", zap.ByteString("body", body))

		return "", "", errors.New("browse get cookie url failure")
	}

	var re *regexp.Regexp
	if re, err = regexp.Compile("{.+?}"); err != nil {
		logger.Error("生成正则表达式失败", zap.Error(err))

		return
	}

	body2 := re.Find(body)
	if body2 == nil {
		logger.Error("匹配正则表达式失败", zap.ByteString("body", body), zap.String("re", re.String()))

		return "", "", errors.New("regexp failure")
	}

	type RailInfo struct {
		Expiration string `json:"exp"`
		DeviceID   string `json:"dfp"`
	}
	railInfo := RailInfo{}
	if err = json.Unmarshal(body2, &railInfo); err != nil {
		logger.Error("解析回调信息错误", zap.Error(err))

		return
	}

	return railInfo.Expiration, railInfo.DeviceID, nil
}

func SetCookie(jar *cookiejar.Jar) (err error) {
	var railExpiration, railDeviceID string
	switch config.Cfg.Login.GetCookieMethod {
	case 1: // 使用 SELENIUM 获取
		if railExpiration, railDeviceID, err = GetBySelenium(jar); err != nil {
			return
		}

	case 2: // 使用 https://kyfw.12306.cn/otn/HttpZF/logdevice 获取
		if railExpiration, railDeviceID, err = GetHttpZF(jar); err != nil {
			return
		}

	case 3: // 自行获取
		railExpiration = config.Cfg.Login.RailExpiration
		railDeviceID = config.Cfg.Login.RailDeviceID

	default:
		return errors.New("config.Cfg.Login.GetCookieMethod error")
	}

	logger.Debug("浏览器设备信息", zap.String("railExpiration", railExpiration), zap.String("railDeviceID", railDeviceID))

	u, _ := url.Parse("https://kyfw.12306.cn")
	jar.SetCookies(u, []*http.Cookie{
		{
			Name:   "RAIL_EXPIRATION",
			Value:  railExpiration,
			Path:   "/",
			Domain: "kyfw.12306.cn",
		},
		{
			Name:   "RAIL_DEVICEID",
			Value:  railDeviceID,
			Path:   "/",
			Domain: "kyfw.12306.cn",
		},
	})

	return
}
