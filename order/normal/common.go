package normal

import (
	"errors"
	"fmt"
	"net/http/cookiejar"
	"strings"
	"time"

	"gogo12306/common"
	"gogo12306/logger"
	"gogo12306/worker"

	"go.uber.org/zap"
)

var globalRepeatSubmitToken string
var ticketInfoForPassengerForm map[string]interface{}

// 以下都是通过 12306 官网源码解析得到
// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js

// 座席类型代号，搜关键字：Q(cP) 或者 aI(cP)
// 9  - 商务座
// TZ - 特等座
// M  - 一等座
// O  - 二等座/二等包座
// 6  - 高级软卧
// 4  - 软卧/一等卧
// F  - 动卧
// 3  - 硬卧/二等卧
// 2  - 软座
// 1  - 硬座
// WZ - 无座

func getPassengerTickets(passengers common.PassengerTicketInfos) string {
	var arr []string
	for _, passenger := range passengers {
		arr = append(arr, fmt.Sprintf("%s,%d,%d,%s,%s,%s,%s,N,%s",
			passenger.SeatType,
			passenger.BedPos,
			passenger.PassengerType,
			passenger.PassengerName,
			passenger.IDTypeCode,
			passenger.IDNumber,
			passenger.MobileNumber,
			passenger.AllEncStr,
		))
	}

	return strings.Join(arr, "_")
}

func getOldPassengers(passengers common.PassengerTicketInfos) (ret string) {
	for _, passenger := range passengers {
		ret += fmt.Sprintf("%s,%s,%s,%d_",
			passenger.PassengerName,
			passenger.IDTypeCode,
			passenger.IDNumber,
			passenger.PassengerType,
		)
	}

	return
}

func DoNormalOrder(jar *cookiejar.Jar, task *worker.Task, leftTicketInfo *common.LeftTicketInfo,
	startDate string, seatIndex int, passengers common.PassengerTicketInfos) (orderID string, err error) {
	if err = SubmitOrder(jar, &SubmitOrderRequest{
		SecretStr:            leftTicketInfo.SecretStr,
		TrainDate:            startDate,
		QueryFromStationName: task.From, // 注意使用中文站名
		QueryToStationName:   task.To,   // 注意使用中文站名
	}); err != nil {
		return
	}

	if err = InitToken(jar); err != nil {
		return
	}

	var (
		ifShowPassCode        bool
		ifShowPassCodeTime    int
		passengerTicketStr    string = getPassengerTickets(passengers)
		oldPassengerTicketStr string = getOldPassengers(passengers)
	)
	if ifShowPassCode, ifShowPassCodeTime, err = CheckOrder(jar, &CheckOrderRequest{
		PassengerTicketStr:    passengerTicketStr,
		OldPassengerTicketStr: oldPassengerTicketStr,
	}); err != nil {
		return
	}

	logger.Debug("是否需要验证码", zap.Bool("ifShowPassCode", ifShowPassCode), zap.Int("ifShowPassCodeTime", ifShowPassCodeTime))

	if err = GetQueueCountResult(jar, &GetQueueCountRequest{
		TrainDate:            startDate,
		TrainNumber:          leftTicketInfo.TrainNumber,
		TrainCode:            leftTicketInfo.TrainCode,
		SeatType:             common.SeatIndexToSeatType(seatIndex),
		QueryFromStationName: task.FromTelegramCode,
		QueryToStationName:   task.ToTelegramCode,
		LeftTicketStr:        leftTicketInfo.LeftTicketStr,
	}); err != nil {
		return
	}

	if ifShowPassCode {
		// TODO 验证码识别

		if ifShowPassCodeTime > 0 {
			time.Sleep(time.Millisecond * time.Duration(ifShowPassCodeTime))
		}
	}

	if err = ConfirmSingleForQueue(jar, &ConfirmSingleForQueueRequest{
		PassengerTicketStr:    passengerTicketStr,
		OldPassengerTicketStr: oldPassengerTicketStr,
		ChooseSeats:           task.ChooseSeats,
		SeatDetailType:        task.SeatDetailType,
	}); err != nil {
		return
	}

	// 每三秒查询一次下单情况
	var (
		retries int
	)
	for {
		retries++
		time.Sleep(time.Second * 3)

		if orderID, err = QueryOrderWaitTime(jar); err != nil {
			return
		} else if orderID != "" {
			break
		} else if retries > 7 {
			logger.Error("已经超过重试上限...")

			return "", errors.New("over retries")
		}
	}

	if orderID == "" {
		return "", errors.New("orderID empty")
	}

	if err = ResultOrderForDcQueue(jar, &ResultOrderForDcQueueRequest{
		OrderID: orderID,
	}); err != nil {
		return
	}

	return
}
