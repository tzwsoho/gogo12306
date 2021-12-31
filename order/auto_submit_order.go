package order

import (
	"bytes"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"go.uber.org/zap"
)

type AutoSubmitOrderRequest struct {
	SecretStr             string // 下单用的密钥
	TrainDate             string // 出发日期
	QueryFromStationName  string // 出发站中文站名
	QueryToStationName    string // 到达站中文站名
	PassengerTicketStr    string
	OldPassengerTicketStr string
}

// AutoSubmitOrder 自动提交订单请求，用于候补票/刷票

// 查询余票网页：https://kyfw.12306.cn/otn/leftTicket/init

// 共有以下参数，搜关键字："undefined" == typeof(submitForm)
// secretStr: 查询余票时每个车次的密钥，只用于下单
// train_date: 乘车日期
// tour_flag: dc 单程, wc 往返，fc 返程，gc 改签

// purpose_codes:
// 参考 cI() 函数
// 需要对查询余票网页上 id 为 sf2 的 radio 的值进行判断，
// 并且根据 login_isDisable 的值会有多种结果

// query_from_station_name: 出发站
// query_to_station_name: 到达站
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
func AutoSubmitOrder(jar *cookiejar.Jar, info *AutoSubmitOrderRequest) (err error) {
	const (
		url0    = "https://%s/otn/confirmPassenger/autoSubmitOrderRequest"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)
	payload := url.Values{}
	payload.Add("secretStr", info.SecretStr)
	payload.Add("train_date", info.TrainDate)
	payload.Add("tour_flag", "dc")
	payload.Add("purpose_codes", passengerTypeToPurposeCodes())
	payload.Add("query_from_station_name", info.QueryFromStationName)
	payload.Add("query_to_station_name", info.QueryToStationName)
	payload.Add("_json_att", "")
	payload.Add("cancel_flag", "2")
	payload.Add("bed_level_order_num", "000000000000000000000000000000")
	payload.Add("passengerTicketStr", info.PassengerTicketStr)
	payload.Add("oldPassengerStr", info.OldPassengerTicketStr)

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	log.Println(req.URL.String())

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

	logger.Debug("自动提交订单", zap.ByteString("body", body))

	return
}
