package login

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"

	"go.uber.org/zap"
)

func GetRailInfo(jar *cookiejar.Jar) (railExpiration, railDeviceID string, err error) {
	var req *http.Request
	req, _ = http.NewRequest("GET", "https://12306-rail-id-v2.pjialin.com/", nil)

	var (
		body []byte
		ok   bool
	)
	if body, ok, _, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("获取浏览器 ID 错误", zap.Error(err))
		return
	} else if !ok {
		logger.Error("获取浏览器 ID 失败", zap.ByteString("body", body))
		return "", "", errors.New("get cookie failure")
	}

	type IDInfo struct {
		ID string `json:"id"`
	}
	idInfo := IDInfo{}
	if err = json.Unmarshal(body, &idInfo); err != nil {
		logger.Error("解析浏览器 ID 信息错误", zap.Error(err))
		return
	}

	var idURL []byte
	if idURL, err = base64.StdEncoding.DecodeString(idInfo.ID); err != nil {
		logger.Error("解析浏览器 ID URL 错误", zap.String("id", idInfo.ID), zap.Error(err))
		return
	}

	logger.Debug("浏览器 ID URL", zap.ByteString("url", idURL))

	req, _ = http.NewRequest("GET", string(idURL), nil)
	httpcli.DefaultHeaders(req)

	if body, ok, _, err = httpcli.DoHttp(req, nil); err != nil {
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

	body = re.Find(body)
	if body == nil {
		logger.Error("匹配正则表达式失败", zap.ByteString("body", body), zap.String("re", re.String()))
		return "", "", errors.New("regexp failure")
	}

	type RailInfo struct {
		Expiration string `json:"exp"`
		DeviceID   string `json:"dfp"`
	}
	railInfo := RailInfo{}
	if err = json.Unmarshal(body, &railInfo); err != nil {
		logger.Error("解析回调信息错误", zap.Error(err))
		return
	}

	return railInfo.Expiration, railInfo.DeviceID, nil
}

func SetCookie(jar *cookiejar.Jar) (err error) {
	var railExpiration, railDeviceID string
	if config.Cfg.Login.GetWebBrowserIDType == 1 { // 使用 pjialin 大佬的网站获取
		if railExpiration, railDeviceID, err = GetRailInfo(jar); err != nil {
			return
		}
	} else if config.Cfg.Login.GetWebBrowserIDType == 2 { // 自行获取
		railExpiration = config.Cfg.Login.RailExpiration
		railDeviceID = config.Cfg.Login.RailDeviceID
	} else {
		return errors.New("config.Cfg.Login.GetWebBrowserIDType error")
	}

	logger.Debug("WebBrowser Cookie Info", zap.String("railExpiration", railExpiration), zap.String("railDeviceID", railDeviceID))

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
