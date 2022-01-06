package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

func InitToken(jar *cookiejar.Jar) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/initDc"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	payload := url.Values{}
	payload.Add("_json_attr", "")

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("获取下单页面信息错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取下单页面信息失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get init dc failure")
	}

	var re1, re2 *regexp.Regexp
	if re1, err = regexp.Compile(`var\s+globalRepeatSubmitToken\s*=\s*(?:'|")([^'"]+?)(?:'|")`); err != nil {
		logger.Error("获取下单页面信息，生成正则表达式 1 失败", zap.Error(err))

		return
	}

	if re2, err = regexp.Compile(`var\s+ticketInfoForPassengerForm\s*=\s*([^;]+?);`); err != nil {
		logger.Error("获取下单页面信息，生成正则表达式 2 失败", zap.Error(err))

		return
	}

	body1 := re1.FindSubmatch(body)
	if body1 == nil || len(body1) != 2 {
		logger.Error("获取下单页面信息，匹配正则表达式 1 失败", zap.ByteString("body", body), zap.String("re", re1.String()))

		return errors.New("regexp 1 match failure")
	}

	body2 := re2.FindSubmatch(body)
	if body2 == nil || len(body2) != 2 {
		logger.Error("获取下单页面信息，匹配正则表达式 2 失败", zap.ByteString("body", body), zap.String("re", re2.String()))

		return errors.New("regexp 2 match failure")
	}

	globalRepeatSubmitToken = string(body1[1])

	var bodyStr2 /*, bodyStr3*/ string
	if bodyStr2, err = url.QueryUnescape(string(body2[1])); err != nil {
		logger.Error("获取下单页面信息，QueryUnescape JSON 2 失败", zap.ByteString("body2", body2[1]), zap.Error(err))

		return
	}

	decoder2 := json.NewDecoder(bytes.NewReader([]byte(strings.ReplaceAll(bodyStr2, "'", "\""))))
	if err = decoder2.Decode(&ticketInfoForPassengerForm); err != nil {
		logger.Error("获取下单页面信息，Decode JSON 2 失败", zap.ByteString("body2", body2[1]), zap.Error(err))

		return errors.New("decode json 2 failure")
	}

	return
}
