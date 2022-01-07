package common

import "strconv"

var LoginIsDisable bool

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js cI() 函数
func PassengerTypeToPurposeCodes() string {
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
