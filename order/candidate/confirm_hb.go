package candidate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type ConfirmHBRequest struct {
	PassengerInfo  string
	CandidateTrain string
	Deadline       int // 截止兑换时间
}

// ConfirmHB 确认候补订单
func ConfirmHB(jar *cookiejar.Jar, info *ConfirmHBRequest) (err error) {
	const (
		url0    = "https://%s/otn/afterNate/confirmHB"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	payload := &url.Values{}
	payload.Add("passengerInfo", info.PassengerInfo)
	payload.Add("jzParam", "")
	payload.Add("hbTrain", info.CandidateTrain)
	payload.Add("lkParam", "") // TODO 接受新增临时客车
	payload.Add("sessionId", "")
	payload.Add("sig", "")
	payload.Add("scene", "nc_login")
	payload.Add("if_receive_wseat", "Y")
	payload.Add("realize_limit_time_diff", strconv.Itoa(info.Deadline))

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
		logger.Error("确认候补订单错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("确认候补订单失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("confirm HB failure")
	}

	logger.Debug("确认候补订单", zap.ByteString("body", body))

	type ConfirmHBData struct {
		Flag bool   `json:"flag,omitempty"`
		Msg  string `json:"msg,omitempty"`
	}

	type ConfirmHBResponse struct {
		Status   bool          `json:"status"`
		Messages []string      `json:"messages"`
		Data     ConfirmHBData `json:"data"`
	}
	response := ConfirmHBResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析确认候补订单返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("确认候补订单失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	} else if !response.Data.Flag {
		logger.Error("确认候补订单失败", zap.ByteString("body", body))

		return errors.New(response.Data.Msg)
	}

	return
}
