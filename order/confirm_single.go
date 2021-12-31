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
	"strings"

	"go.uber.org/zap"
)

type ConfirmSingleForQueueRequest struct {
	PassengerTicketStr    string
	OldPassengerTicketStr string
	ChooseSeats           []string
	SeatDetailType        []string
}

// ConfirmSingleForQueue 确认排队情况
func ConfirmSingleForQueue(jar *cookiejar.Jar, info *ConfirmSingleForQueueRequest) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/confirmSingleForQueue"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)

	payload := &url.Values{}
	payload.Add("passengerTicketStr", info.PassengerTicketStr)
	payload.Add("oldPassengerStr", info.OldPassengerTicketStr)
	payload.Add("randCode", "")
	payload.Add("purpose_codes", ticketInfoForPassengerForm["purpose_codes"].(string))
	payload.Add("key_check_isChange", ticketInfoForPassengerForm["key_check_isChange"].(string))

	// payload.Add("leftTicket", info.LeftTicketStr)
	payload.Add("leftTicketStr", ticketInfoForPassengerForm["leftTicketStr"].(string))
	// payload.Add("leftTicket", ticketInfoForPassengerForm["queryLeftTicketRequestDTO"].(map[string]interface{})["ypInfoDetail"].(string))

	payload.Add("train_location", ticketInfoForPassengerForm["train_location"].(string))

	payload.Add("choose_seats", strings.Join(info.ChooseSeats, ""))
	if len(info.SeatDetailType) > 0 {
		payload.Add("seatDetailType", "000")
	} else {
		payload.Add("seatDetailType", strings.Join(info.SeatDetailType, ""))
	}

	payload.Add("is_jy", "N")
	payload.Add("is_cj", "Y")

	payload.Add("encryptedData", "")
	payload.Add("whatsSelect", "1")
	payload.Add("roomType", "00") // TODO
	payload.Add("dwAll", "N")     // TODO

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
		logger.Error("确认排队情况错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("确认排队情况失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get queue count failure")
	}

	logger.Debug("确认排队情况", zap.ByteString("body", body))

	type ConfirmSingleForQueueData struct {
		SubmitStatus bool `json:"submitStatus"`
	}

	type ConfirmSingleForQueueResponse struct {
		Status   bool                      `json:"status"`
		Messages []string                  `json:"messages"`
		Data     ConfirmSingleForQueueData `json:"data"`
	}
	response := ConfirmSingleForQueueResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析确认排队情况返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("确认排队情况失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	}

	return
}
