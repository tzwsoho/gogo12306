package common

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
