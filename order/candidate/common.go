package candidate

import (
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"gogo12306/common"
	"gogo12306/worker"
)

type CandidateInfo struct {
	Info      string // 候补人数情况
	Deadline  string // 截止兑换日期时间
	ReserveNo string // 候补订单号
}

func getCandidateSecretStr(secretStr string, seatIndex int) string {
	return url.QueryEscape(secretStr) + "#" + common.SeatIndexToSeatType(seatIndex) + "|"
}

// getConfirmHBSecret 确认候补订单发送的密钥串，这个函数解析时有点复杂
func getConfirmHBSecret(passengers common.PassengerTicketInfos) (ret string) {
	// 先获取乘客信息 passengerInfo
	// https://kyfw.12306.cn/otn/view/lineUp_toPay.html
	// passengerInfo = '<%= (obj.passenger_type || "1") + '#' + obj.passenger_name + '#' + obj.passenger_id_type_code + '#' +
	// obj.passenger_id_no + '#' + obj.isOldThan60 + '#' + obj.total_times +'#' + obj.allEncStr +'#'%>'>

	// Y 为 passengerInfo 数组：
	// https://kyfw.12306.cn/otn/personalJS/dist/lineUp_toPay/main_v11024.js 关键词：Y.push(h)

	// confirmHB 函数为 F(r) 所调用，其中 r 是 Y 元素拼凑而成：
	// o = Y[a].split("#"), l = "Y" == o[4] && _ > 0 ? 1 : 0, s = o[0] + "#" + o[1] + "#" + o[2] + "#" + o[3] + "#" + o[6] + "#" + l, r.push(s)
	// o[4]: isOldThan60
	// o[0]: passenger_type
	// o[1]: passenger_name
	// o[2]: passenger_id_type_code
	// o[3]: passenger_id_no
	// o[6]: allEncStr
	// l: 乘客 60 岁以上，并且有选择卧铺座席类型，l 的值是 1，否则是 0

	// 下划线变量是卧铺（硬卧/二等卧，动卧，软卧/一等卧，高级软卧，A/I/J 暂时没找到对应的座席类型）的数量：
	// "3" != n.data.hbTrainList[s].seat_type_code &&
	// "F" != n.data.hbTrainList[s].seat_type_code &&
	// "4" != n.data.hbTrainList[s].seat_type_code &&
	// "6" != n.data.hbTrainList[s].seat_type_code &&
	// "A" != n.data.hbTrainList[s].seat_type_code &&
	// "I" != n.data.hbTrainList[s].seat_type_code &&
	// "J" != n.data.hbTrainList[s].seat_type_code || _++

	for _, passenger := range passengers {
		var l int
		switch passenger.SeatType {
		case "3", "F", "4", "6", "A", "I", "J":
			if passenger.IsOlderThan60 == "Y" {
				l = 1
			}
		}

		ret += fmt.Sprintf("%d#%s#%s#%s#%s#%d;",
			passenger.PassengerType,
			passenger.PassengerName,
			passenger.IDTypeCode,
			passenger.IDNumber,
			passenger.AllEncStr,
			l,
		)
	}

	return
}

func getCandidateTrains(trainNos []string, seatIndex int) (ret string) {
	// https://kyfw.12306.cn/otn/personalJS/dist/lineUp_toPay/main_v11024.js 关键词：data-trainno
	for _, trainNo := range trainNos {
		ret += fmt.Sprintf("%s,%s#", trainNo, common.SeatIndexToSeatType(seatIndex))
	}

	return
}

func DoCandidate(jar *cookiejar.Jar, task *worker.Task, leftTicketInfo *common.LeftTicketInfo,
	seatIndex int, passengers common.PassengerTicketInfos) (info *CandidateInfo, err error) {
	info = &CandidateInfo{}

	var secretStr string = getCandidateSecretStr(leftTicketInfo.SecretStr, seatIndex)

	if err = CheckFace(jar, &CheckFaceRequest{
		SecretStr: secretStr,
	}); err != nil {
		return
	}

	var trainNos []string
	if trainNos, info.Info, err = GetSuccessRate(jar, &GetSuccessRateRequest{
		SecretStr: strings.TrimSuffix(secretStr, "|"),
	}); err != nil {
		return
	}

	if err = SubmitOrder(jar, &SubmitOrderRequest{
		SecretStr: secretStr,
	}); err != nil {
		return
	}

	if info.Deadline, err = PassengerInitAPI(jar, &PassengerInitAPIRequest{}); err != nil {
		return
	}

	if err = GetQueueNum(jar, &GetQueueNumRequest{}); err != nil {
		return
	}

	if info.ReserveNo, err = ConfirmHB(jar, &ConfirmHBRequest{
		PassengerInfo:  getConfirmHBSecret(passengers),
		CandidateTrain: getCandidateTrains(trainNos, seatIndex),
		Deadline:       task.CandidateDeadline,
	}); err != nil {
		return
	}

	return
}
