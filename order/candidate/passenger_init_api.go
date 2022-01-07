package candidate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type PassengerInitAPIRequest struct {
}

// PassengerInitAPI 候补结果
func PassengerInitAPI(jar *cookiejar.Jar, request *PassengerInitAPIRequest) (deadline string, err error) {
	const (
		url0    = "https://%s/otn/afterNate/passengerInitApi"
		referer = "https://kyfw.12306.cn/otn/view/lineUp_toPay.html"
	)

	buf := bytes.NewBuffer([]byte{})
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("候补结果错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("候补结果失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return "", errors.New("passenger init api failure")
	}

	logger.Debug("候补结果", zap.ByteString("body", body))

	type PassengerInitAPIData struct {
		JZDHDateS      string   `json:"jzdhDateS"` // jzdh = 截止兑换
		JZDHHourS      string   `json:"jzdhHourS"`
		JZDHDateE      string   `json:"jzdhDateE"`
		JZDHHourE      string   `json:"jzdhHourE"`
		JZDHDiffSelect []string `json:"jzdhDiffSelect"` // 可选的截止兑换时间（单位：分钟，默认：360）
	}

	type PassengerInitAPIResponse struct {
		Status   bool                 `json:"status"`
		Messages []string             `json:"messages"`
		Data     PassengerInitAPIData `json:"data"`
	}
	response := PassengerInitAPIResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析候补结果返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("候补结果失败", zap.Strings("错误消息", response.Messages))

		return "", errors.New(strings.Join(response.Messages, ""))
	}

	deadline = response.Data.JZDHDateE + " " + response.Data.JZDHHourE

	logger.Info("候补结果",
		zap.String("候补开始日期", response.Data.JZDHDateS+" "+response.Data.JZDHHourS),
		zap.String("截止兑换日期", deadline),
		zap.Strings("可选的截止兑换时间", response.Data.JZDHDiffSelect),
	)

	return
}
