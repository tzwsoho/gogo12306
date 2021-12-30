package order

import (
	"fmt"
	"gogo12306/worker"
	"strings"
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

func GetPassengerTicketsForAutoSubmit(passengers worker.PassengerTicketInfos) string {
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

func GetOldPassengersForAutoSubmit(passengers worker.PassengerTicketInfos) (ret string) {
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

func GetPassengerTickets(passengers worker.PassengerTicketInfos) string {
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

func GetOldPassengers(passengers worker.PassengerTicketInfos) (ret string) {
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
