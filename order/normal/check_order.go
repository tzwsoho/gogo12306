package normal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type CheckOrderRequest struct {
	PassengerTicketStr    string
	OldPassengerTicketStr string
}

// CheckOrder 下单成功后检查订单信息
// ifShowPassCode: 如果此字段存在，并且值为 Y，则需要做验证码识别
// ifShowPassCodeTime: 验证码识别完之前要等待的毫秒数
func CheckOrder(jar *cookiejar.Jar, info *CheckOrderRequest) (ifShowPassCode bool, ifShowPassCodeTime int, err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/checkOrderInfo"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)

	payload := &url.Values{}
	payload.Add("cancel_flag", "2")
	payload.Add("bed_level_order_num", "000000000000000000000000000000")
	payload.Add("passengerTicketStr", info.PassengerTicketStr)
	payload.Add("oldPassengerStr", info.OldPassengerTicketStr)
	payload.Add("tour_flag", "dc")
	payload.Add("randCode", "")
	payload.Add("whatsSelect", "1")
	payload.Add("sessionId", "")
	payload.Add("sig", "")
	payload.Add("scene", "nc_login")
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

		return false, 0, errors.New("check order info failure")
	}

	logger.Debug("检查订单信息", zap.ByteString("body", body))

	type CheckOrderData struct {
		SubmitStatus       bool   `json:"submitStatus"`
		IfShowPassCodeTime int    `json:"ifShowPassCodeTime,string"`
		IfShowPassCode     string `json:"ifShowPassCode,omitempty"`
	}

	type CheckOrderResponse struct {
		Status   bool           `json:"status"`
		Messages []string       `json:"messages"`
		Data     CheckOrderData `json:"data"`
	}
	response := CheckOrderResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析检查订单信息返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("检查订单信息失败", zap.Strings("错误消息", response.Messages))

		return false, 0, errors.New(strings.Join(response.Messages, ""))
	} else if !response.Data.SubmitStatus {
		logger.Error("检查订单信息结果: 提交失败", zap.Strings("错误消息", response.Messages))

		return false, 0, errors.New(strings.Join(response.Messages, ""))
	} else {
		logger.Info("检查订单信息结果: 提交成功，尝试排队...")

		ifShowPassCode = (response.Data.IfShowPassCode == "Y")
		ifShowPassCodeTime = response.Data.IfShowPassCodeTime
	}

	return
}
