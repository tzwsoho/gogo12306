package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

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

	var re1, re2, re3 *regexp.Regexp
	if re1, err = regexp.Compile("var\\s+globalRepeatSubmitToken\\s*=\\s*(?:'|\")([^'\"]+?)(?:'|\")"); err != nil {
		logger.Error("生成正则表达式 1 失败", zap.Error(err))

		return
	}

	if re2, err = regexp.Compile("var\\s+ticketInfoForPassengerForm\\s*=\\s*([^;]+?);"); err != nil {
		logger.Error("生成正则表达式 2 失败", zap.Error(err))

		return
	}

	if re3, err = regexp.Compile("var\\s+orderRequestDTO\\s*=\\s*([^;]+?);"); err != nil {
		logger.Error("生成正则表达式 3 失败", zap.Error(err))

		return
	}

	body1 := re1.FindSubmatch(body)
	if body1 == nil || len(body1) != 2 {
		logger.Error("匹配正则表达式 1 失败", zap.ByteString("body", body), zap.String("re", re1.String()))

		return errors.New("regexp 1 match failure")
	}

	body2 := re2.FindSubmatch(body)
	if body2 == nil || len(body2) != 2 {
		logger.Error("匹配正则表达式 2 失败", zap.ByteString("body", body), zap.String("re", re2.String()))

		return errors.New("regexp 2 match failure")
	}

	body3 := re3.FindSubmatch(body)
	if body3 == nil || len(body3) != 2 {
		logger.Error("匹配正则表达式 3 失败", zap.ByteString("body", body), zap.String("re", re3.String()))

		return errors.New("regexp 3 match failure")
	}

	globalRepeatSubmitToken = string(body1[1])

	var bodyStr2, bodyStr3 string
	if bodyStr2, err = url.QueryUnescape(string(body2[1])); err != nil {
		logger.Error("QueryUnescape JSON 2 失败", zap.ByteString("body2", body2[1]), zap.Error(err))

		return
	}

	if bodyStr3, err = url.QueryUnescape(string(body3[1])); err != nil {
		logger.Error("QueryUnescape JSON 3 失败", zap.ByteString("body3", body3[1]), zap.Error(err))

		return
	}

	decoder2 := json.NewDecoder(bytes.NewReader([]byte(strings.ReplaceAll(bodyStr2, "'", "\""))))
	if err = decoder2.Decode(&ticketInfoForPassengerForm); err != nil {
		logger.Error("Decode JSON 2 失败", zap.ByteString("body2", body2[1]), zap.Error(err))

		return errors.New("decode json 2 failure")
	}

	decoder3 := json.NewDecoder(bytes.NewReader([]byte(strings.ReplaceAll(bodyStr3, "'", "\""))))
	if err = decoder3.Decode(&orderRequestDTO); err != nil {
		logger.Error("Decode JSON 3 失败", zap.ByteString("body3", body3[1]), zap.Error(err))

		return errors.New("decode json 3 failure")
	}

	return
}
