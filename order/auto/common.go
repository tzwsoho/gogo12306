package auto

import (
	"fmt"
	"gogo12306/common"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http/cookiejar"
	"strings"

	"go.uber.org/zap"
)

func GetPassengerTicketsForAutoSubmit(passengers common.PassengerTicketInfos) string {
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

func GetOldPassengersForAutoSubmit(passengers common.PassengerTicketInfos) (ret string) {
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

func DoAutoOrder(jar *cookiejar.Jar, task *worker.Task, leftTicketInfo *common.LeftTicketInfo,
	startDate string, passengers common.PassengerTicketInfos) (orderID string, err error) {
	var (
		passengerTicketStr    string = GetPassengerTicketsForAutoSubmit(passengers)
		oldPassengerTicketStr string = GetOldPassengersForAutoSubmit(passengers)
	)
	if orderID, err = AutoSubmitOrder(jar, &AutoSubmitOrderRequest{
		SecretStr:             leftTicketInfo.SecretStr,
		TrainDate:             startDate,
		QueryFromStationName:  task.From,
		QueryToStationName:    task.To,
		PassengerTicketStr:    passengerTicketStr,
		OldPassengerTicketStr: oldPassengerTicketStr,
	}); err != nil {
		logger.Error("自动下单失败", zap.Error(err))

		return
	}

	return
}
