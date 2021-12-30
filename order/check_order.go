package order

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

type CheckOrderInfo struct {
	Passengers []*worker.PassengerTicketInfo // 乘客列表
}

// CheckOrder 下单成功后检查订单信息
func CheckOrder(jar *cookiejar.Jar, info *CheckOrderInfo) (ifShowPassCodeTime int, ifShowPassCode bool, err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/checkOrderInfo"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)

	payload := &url.Values{}
	payload.Add("cancel_flag", "2")
	payload.Add("bed_level_order_num", "000000000000000000000000000000")
	payload.Add("passengerTicketStr", getPassengerTickets(info.Passengers))
	payload.Add("oldPassengerStr", getOldPassengers(info.Passengers))
	payload.Add("tour_flag", "dc")
	payload.Add("randCode", "")
	payload.Add("_json_att", "")
	payload.Add("REPEAT_SUBMIT_TOKEN", globalRepeatSubmitToken)

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
		logger.Error("检查订单信息错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("检查订单信息失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return 0, false, errors.New("check order info failure")
	}

	logger.Info("检查订单信息", zap.ByteString("body", body))

	type CheckOrderData struct {
		SubmitStatus       bool   `json:"submitStatus"`
		IfShowPassCodeTime int    `json:"ifShowPassCodeTime,string"`
		IfShowPassCode     string `json:"ifShowPassCode,omitempty"`
	}

	type CheckOrderInfoResult struct {
		Status   bool           `json:"status"`
		Messages []string       `json:"messages"`
		Data     CheckOrderData `json:"data"`
	}
	result := CheckOrderInfoResult{}
	if err = json.Unmarshal(body, &result); err != nil {
		logger.Error("解析检查订单信息返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !result.Status {
		logger.Error("检查订单信息失败", zap.Strings("错误消息", result.Messages))

		return 0, false, errors.New(strings.Join(result.Messages, ""))
	} else if !result.Data.SubmitStatus {
		logger.Error("检查订单信息结果: 提交失败", zap.Strings("错误消息", result.Messages))

		return 0, false, errors.New(strings.Join(result.Messages, ""))
	} else {
		logger.Info("检查订单信息结果: 提交成功，尝试排队...")

		ifShowPassCodeTime = result.Data.IfShowPassCodeTime
		ifShowPassCode = (result.Data.IfShowPassCode == "Y")
	}

	return
}
