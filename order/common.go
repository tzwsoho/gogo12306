package order

import (
	"errors"
	"fmt"
	"gogo12306/logger"
	"gogo12306/login"
	"gogo12306/notifier"
	"gogo12306/worker"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

var LoginIsDisable bool
var CheckUserMDID string

var globalRepeatSubmitToken string
var ticketInfoForPassengerForm map[string]interface{}
var orderRequestDTO map[string]interface{}

// 以下都是通过 12306 官网源码解析得到
// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js

// 坐席类型代号，搜关键字：Q(cP) 或者 aI(cP)
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

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js cI() 函数
func passengerTypeToPurposeCodes() string {
	// if 点选 “学生票” {
	// 	if LoginIsDisable {
	// 		return "0X1C"
	// 	} else {
	// 		return "0X00"
	// 	}
	// } else {
	if LoginIsDisable {
		return "1C"
	} else {
		return "ADULT"
	}
	// }
}

func getPassengerTicketsForAutoSubmit(passengers worker.PassengerTicketInfos) string {
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

func getOldPassengersForAutoSubmit(passengers worker.PassengerTicketInfos) (ret string) {
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

func getPassengerTickets(passengers worker.PassengerTicketInfos) string {
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

func getOldPassengers(passengers worker.PassengerTicketInfos) (ret string) {
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

// 0 - 9 - 商务座
// 1 - TZ - 特等座
// 2 - M - 一等座
// 3 - O - 二等座/二等包座
// 4 - 6 - 高级软卧
// 5 - 4 - 软卧/一等卧
// 6 - F - 动卧
// 7 - 3 - 硬卧/二等卧
// 8 - 2 - 软座
// 9 - 1 - 硬座
// 10 - WZ - 无座
// 11 - 其他
func SeatIndexToSeatType(seatIndex int) string {
	var seatTypes []string = []string{"9", "TZ", "M", "O", "6", "4", "F", "3", "2", "1", "WZ", ""}
	if seatIndex < 0 || seatIndex >= len(seatTypes) {
		return ""
	}

	return seatTypes[seatIndex]
}

func SeatIndexToSeatName(seatIndex int) string {
	var seatNames []string = []string{
		"商务座", "特等座", "一等座", "二等座/二等包座",
		"高级软卧", "软卧/一等卧", "动卧", "硬卧/二等卧",
		"软座", "硬座", "无座", "其他"}
	if seatIndex < 0 || seatIndex >= len(seatNames) {
		return strconv.Itoa(seatIndex)
	}

	return seatNames[seatIndex]
}

func DoOrder(jar *cookiejar.Jar, task *worker.Task, leftTicketInfo *worker.LeftTicketInfo,
	startDate, trainCode string, seatIndex int, passengers worker.PassengerTicketInfos) (err error) {
	if err = login.CheckAndRelogin(jar); err != nil {
		return
	}

	// TODO 判断小黑屋

	// TODO 候补算法逻辑：
	// ①本车次无余票但可候补，马上进行候补
	// ②本车次无余票但可候补，不进行候补，待所有车次都查询完均无余票时，再遍历选择的车次并进行候补

	// TODO 有部分坐席类型不能进行候补，需要判断

	if task.OrderType == 1 { // 普通购票
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
			SeatType:             SeatIndexToSeatType(seatIndex),
			QueryFromStationName: task.FromTelegramCode,
			QueryToStationName:   task.ToTelegramCode,
			LeftTicketStr:        leftTicketInfo.LeftTicketStr,
		}); err != nil {
			return
		}

		if ifShowPassCode {
			// TODO 验证码识别
		}

		if ifShowPassCodeTime > 0 {
			time.Sleep(time.Millisecond * time.Duration(ifShowPassCodeTime))
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
		var orderID string
		for {
			if orderID, err = QueryOrderWaitTime(jar); err != nil || orderID != "" {
				break
			}

			time.Sleep(time.Second * 3)
		}

		if orderID == "" {
			return
		}

		if err = ResultOrderForDcQueue(jar, &ResultOrderForDcQueueRequest{
			OrderID: orderID,
		}); err != nil {
			return
		}

		notifier.Broadcast(fmt.Sprintf("GOGO12306 已成功帮您抢到 %s 至 %s，出发时间 %s %s，车次 %s，乘客: %s 的车票，订单号为 %s，请尽快登陆 12306 网站完成购票支付",
			task.From, task.To, startDate, leftTicketInfo.StartTime, leftTicketInfo.TrainCode, passengers.Names(), orderID,
		))

		task.Done <- struct{}{}
		return
	} else if task.OrderType == 2 { // 自动捡漏下单
		var (
			passengerTicketStr    string = getPassengerTicketsForAutoSubmit(passengers)
			oldPassengerTicketStr string = getOldPassengersForAutoSubmit(passengers)
		)
		if err = AutoSubmitOrder(jar, &AutoSubmitOrderRequest{
			SecretStr:             leftTicketInfo.SecretStr,
			TrainDate:             startDate,
			QueryFromStationName:  task.FromTelegramCode, // 注意使用电报码
			QueryToStationName:    task.ToTelegramCode,   // 注意使用电报码
			PassengerTicketStr:    passengerTicketStr,
			OldPassengerTicketStr: oldPassengerTicketStr,
		}); err != nil {
			logger.Error("自动下单失败", zap.Error(err))

			return
		}
	} else if task.OrderCandidate { // 候补票
		// TODO
	} else { // 不接受候补
		logger.Debug("由于设置不接受候补，忽略此车次和坐席...",
			zap.String("车次", trainCode),
			zap.String("坐席类型", SeatIndexToSeatName(seatIndex)),
		)

		return errors.New("candicate not allow")
	}

	return
}
