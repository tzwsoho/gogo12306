package normal

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
	"time"

	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"

	"go.uber.org/zap"
)

type GetQueueCountRequest struct {
	TrainDate            string
	TrainNumber          string
	TrainCode            string
	SeatType             string
	QueryFromStationName string // 出发站电报码
	QueryToStationName   string // 到达站电报码
	LeftTicketStr        string // 余票密钥串
}

// GetQueueCountResult 获取排队信息
func GetQueueCountResult(jar *cookiejar.Jar, request *GetQueueCountRequest) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/getQueueCount"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)

	var trainDate time.Time
	if trainDate, err = time.Parse("2006-01-02", request.TrainDate); err != nil {
		logger.Error("获取排队信息，解析出发日期错误", zap.String("trainDate", request.TrainDate), zap.Error(err))

		return
	}

	payload := &url.Values{}
	payload.Add("train_date", trainDate.Format("Mon Jan 01 2006 00:00:00 GMT-0700 (中国标准时间)"))

	payload.Add("train_no", request.TrainNumber)
	payload.Add("stationTrainCode", request.TrainCode)
	// payload.Add("train_no", ticketInfoForPassengerForm["queryLeftTicketRequestDTO"].(map[string]interface{})["train_no"].(string))
	// payload.Add("stationTrainCode", ticketInfoForPassengerForm["queryLeftTicketRequestDTO"].(map[string]interface{})["station_train_code"].(string))

	payload.Add("seatType", request.SeatType)

	payload.Add("fromStationTelecode", request.QueryFromStationName)
	payload.Add("toStationTelecode", request.QueryToStationName)
	// payload.Add("fromStationTelecode", ticketInfoForPassengerForm["queryLeftTicketRequestDTO"].(map[string]interface{})["from_station"].(string))
	// payload.Add("toStationTelecode", ticketInfoForPassengerForm["queryLeftTicketRequestDTO"].(map[string]interface{})["to_station"].(string))

	// payload.Add("leftTicket", request.LeftTicketStr)
	// payload.Add("leftTicket", ticketInfoForPassengerForm["leftTicketStr"].(string))
	payload.Add("leftTicket", ticketInfoForPassengerForm["queryLeftTicketRequestDTO"].(map[string]interface{})["ypInfoDetail"].(string))

	payload.Add("purpose_codes", ticketInfoForPassengerForm["purpose_codes"].(string))
	payload.Add("train_location", ticketInfoForPassengerForm["train_location"].(string))
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
		logger.Error("获取排队信息错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取排队信息失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get queue count failure")
	}

	logger.Debug("获取排队信息", zap.ByteString("body", body))

	type GetQueueCountData struct {
		Count  int    `json:"count,string"`
		CountT int    `json:"countT,string"`
		Ticket string `json:"ticket"`
		Op1    bool   `json:"op_1,string"`
		Op2    bool   `json:"op_2,string"`
	}

	type GetQueueCountResponse struct {
		Status   bool              `json:"status"`
		Messages []string          `json:"messages"`
		Data     GetQueueCountData `json:"data"`
	}
	response := GetQueueCountResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析排队信息返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("获取排队信息失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	}

	leftTickets := int64(0)
	for _, ticket := range strings.Split(response.Data.Ticket, ",") {
		var leftTicket int64
		if leftTicket, err = strconv.ParseInt(ticket, 10, 64); err != nil {
			logger.Error("获取排队信息，转换票数错误",
				zap.String("ticket", ticket),
				zap.Error(err),
			)

			continue
		}

		leftTickets += leftTicket
	}

	if leftTickets == 0 {
		logger.Warn("排队失败，余票不足!!!")

		return errors.New("get queue count left ticket not enough")
	}

	logger.Info("排队成功...",
		zap.Int("前面等待人数", response.Data.CountT),
		zap.Int64("剩余票数", leftTickets),
	)

	return
}
