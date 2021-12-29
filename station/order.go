package station

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

type AutoOrderInfo struct {
	SecretStr            string                        // 下单用的密钥
	TrainDate            string                        // 出发日期
	QueryFromStationName string                        // 出发站电报码
	QueryToStationName   string                        // 到达站电报码
	Passengers           []*worker.PassengerTicketInfo // 乘客列表
}

type OrderInfo struct {
	SecretStr            string // 下单用的密钥
	TrainDate            string // 出发日期
	QueryFromStationName string // 出发站电报码
	QueryToStationName   string // 到达站电报码
}

// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js cI() 函数
func passengerTypeToPurposeCodes() string {
	// if 点选 “学生票” {
	// 	if loginIsDisable {
	// 		return "0X1C"
	// 	} else {
	// 		return "0X00"
	// 	}
	// } else {
	if loginIsDisable {
		return "1C"
	} else {
		return "ADULT"
	}
	// }
}

func getPassengerTicketsForAutoSubmit(passengers []*worker.PassengerTicketInfo) string {
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

func getOldPassengersForAutoSubmit(passengers []*worker.PassengerTicketInfo) (ret string) {
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

// 查询余票网页：https://kyfw.12306.cn/otn/leftTicket/init

// AutoOrderTicket 自动提交订单请求，用于候补票/刷票
// 共有以下参数，搜关键字："undefined" == typeof(submitForm)
// secretStr: 查询余票时每个车次的密钥，只用于下单
// train_date: 乘车日期
// tour_flag: dc 单程, wc 往返

// purpose_codes:
// 参考 cI() 函数
// 需要对查询余票网页上 id 为 sf2 的 radio 的值进行判断，
// 并且根据 login_isDisable 的值会有多种结果

// query_from_station_name: 出发站代号
// query_to_station_name: 到达站代号
// _json_att: 貌似没用
// cancel_flag: 固定是 2
// bed_level_order_num: 固定是 000000000000000000000000000000
// passengerTicketStr: getpassengerTicketsForAutoSubmit() 函数的返回值
// oldPassengerStr: getOldPassengersForAutoSubmit() 函数的返回值

// getpassengerTicketsForAutoSubmit 函数，返回乘客信息列表
// 乘客信息包含以下内容，用英文逗号 , 隔开，每个乘客之间用下划线 _ 隔开:
// 座席类型代号,卧铺位置,车票类型,乘客姓名,乘客证件类型,乘客证件号码,乘客手机号,保存常用联系人(Y/N),乘客联系人加密字符串
// 座席类型代号：参见上面的注释
// 卧铺位置: 0 - 不限，3 - 上铺，2 - 中铺，1 - 下铺（查询余票网页 id 为 ticketype_ 的 select 组件）
// 车票类型：1 - 成人票，2 - 儿童票，3 - 学生票，4 - 残军票（搜关键字：cV == "1"）
// 证件类型:
// https://kyfw.12306.cn/otn/resources/merged/passengerInfo_js.js，搜关键字: passenger_card_type
// 1 - 二代身份证，2 - 一代身份证，3 - 临时身份证，B - 护照，H - 外国人居留证，C - 港澳居民来往内地通行证，G - 台湾居民来往大陆通行证

// getOldPassengersForAutoSubmit 函数，返回旧乘客信息列表
// 旧乘客信息包含以下内容，用英文逗号 , 隔开，每个乘客之间用下划线 _ 隔开:
// 乘客姓名,乘客证件类型,乘客证件号码,乘客类型
// 乘客类型与 getpassengerTicketsForAutoSubmit 中的 车票类型 意义一致（参照 bv 函数）
func AutoOrderTicket(jar *cookiejar.Jar, info *AutoOrderInfo) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/autoSubmitOrderRequest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	secretStr := url.QueryEscape(info.SecretStr)
	trainDate := url.QueryEscape(info.TrainDate)
	purposeCodes := url.QueryEscape(passengerTypeToPurposeCodes())
	passengerTicketStr := url.QueryEscape(getPassengerTicketsForAutoSubmit(info.Passengers))
	oldPassengerStr := url.QueryEscape(getOldPassengersForAutoSubmit(info.Passengers))

	payload := "secretStr=" + secretStr +
		"&train_date=" + trainDate +
		"&tour_flag=dc" +
		"&purpose_codes=" + purposeCodes +
		"&query_from_station_name=" + info.QueryFromStationName +
		"&query_to_station_name=" + info.QueryToStationName +
		"&_json_att=" +
		"&cancel_flag=2" +
		"&bed_level_order_num=000000000000000000000000000000" +
		"&passengerTicketStr=" + passengerTicketStr +
		"&oldPassengerStr=" + oldPassengerStr
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("自动提交订单错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("自动提交订单失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("auto submit order failure")
	}

	logger.Info("自动提交订单", zap.ByteString("body", body))

	return
}

// OrderTicket 一般下单请求，用于普通购票
func OrderTicket(jar *cookiejar.Jar, info *OrderInfo) (err error) {
	const (
		url0    = "https://%s/otn/leftTicket/submitOrderRequest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	secretStr := url.QueryEscape(info.SecretStr)
	trainDate := url.QueryEscape(info.TrainDate)
	purposeCodes := url.QueryEscape(passengerTypeToPurposeCodes())

	checkusermdId := ""
	if len(checkUserMDID) > 0 {
		checkusermdId = "_json_att=" + url.QueryEscape(checkUserMDID)
	}

	payload := "secretStr=" + secretStr +
		"&train_date=" + trainDate +
		"&back_train_date=" + url.QueryEscape(time.Now().Format("2006-01-02")) +
		"&tour_flag=gc" +
		"&purpose_codes=" + purposeCodes +
		"&query_from_station_name=" + info.QueryFromStationName +
		"&query_to_station_name=" + info.QueryToStationName +
		"&" + checkusermdId
	buf := bytes.NewBuffer([]byte(payload))

	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("提交订单错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("提交订单失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("submit order failure")
	}

	logger.Info("提交订单", zap.ByteString("body", body))

	return
}

// CheckOrderInfo 下单成功后检查订单信息
func CheckOrderInfo(jar *cookiejar.Jar, info *OrderInfo) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/checkOrderInfo"
		referer = "https://kyfw.12306.cn/otn/confirmPassenger/initDc"
	)

	type CheckOrderInfoRequest struct {
		CancelFlag         string `json:"cancel_flag"`
		BedLevelOrderNum   string `json:"bed_level_order_num"`
		PassengerTicketStr string `json:"passengerTicketStr"`
		OldPassengerStr    string `json:"oldPassengerStr"`
		TourFlag           string `json:"tour_flag"`
		RandCode           string `json:"randCode"`
		WhatsSelect        string `json:"whatsSelect"`
		SessionId          string `json:"sessionId"`
		Sig                string `json:"sig"`
		Scene              string `json:"scene"`
	}

	coiReq := &CheckOrderInfoRequest{
		CancelFlag:         "2",
		BedLevelOrderNum:   "000000000000000000000000000000",
		PassengerTicketStr: "", // TODO
		OldPassengerStr:    "", // TODO
		TourFlag:           "", //TODO
		RandCode:           "", // TODO
		WhatsSelect:        "", // TODO
		SessionId:          "", // TODO
		Sig:                "", // TODO
		Scene:              "nc_login",
	}

	var payload []byte
	if payload, err = json.Marshal(coiReq); err != nil {
		logger.Error("序列化请求错误", zap.Any("coiReq", coiReq), zap.Error(err))
		return
	}

	buf := bytes.NewBuffer(payload)
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("检查订单信息错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("检查订单信息失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("check order info failure")
	}

	logger.Info("检查订单信息", zap.ByteString("body", body))

	return
}
