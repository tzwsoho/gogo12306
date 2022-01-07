package common

import "strconv"

// 0 - 9 - 商务座
// 1 - TZ - 特等座
// 2 - M - 一等座
// 3 - O - 二等座
// 4 - 6 - 高级软卧
// 5 - 4 - 软卧
// 6 - F - 动卧
// 7 - 3 - 硬卧
// 8 - 2 - 软座
// 9 - 1 - 硬座
// 10 - WZ - 无座
// 11 -  - 其他
// 12 - 5 - 包厢硬卧
// 13 - 7 - 一等软座
// 14 - 8 - 二等软座
// 15 - A - 高级动卧
// 16 - B - 混编硬座
// 17 - C - 混编硬卧
// 18 - E - 特等软座
// 19 - H - 一人软包
// 20 - I - 一等卧
// 21 - J - 二等卧
// 22 - K - 混编软座
// 23 - L - 混编软卧
// 24 - P - 特等座
// 25 - Q - 多功能座
// 26 - S - 二等包座
func SeatIndexToSeatType(seatIndex int) string {
	var seatTypes []string = []string{
		"9", "TZ", "M", "O", "6", "4", "F", "3", "2", "1",
		"WZ", "", "5", "7", "8", "A", "B", "C", "E", "H",
		"I", "J", "K", "L", "P", "Q", "S",
	}
	if seatIndex < 0 || seatIndex >= len(seatTypes) {
		return ""
	}

	return seatTypes[seatIndex]
}

func SeatIndexToSeatName(seatIndex int) string {
	var seatNames []string = []string{
		"商务座", "特等座", "一等座", "二等座",
		"高级软卧", "软卧", "动卧", "硬卧",
		"软座", "硬座", "无座", "其他",
		"包厢硬卧", "一等软座", "二等软座", "高级动卧",
		"混编硬座", "混编硬卧", "特等软座", "一人软包",
		"一等卧", "二等卧", "混编软座", "混编软卧",
		"特等座", "多功能座", "二等包座",
	}
	if seatIndex < 0 || seatIndex >= len(seatNames) {
		return strconv.Itoa(seatIndex)
	}

	return seatNames[seatIndex]
}
