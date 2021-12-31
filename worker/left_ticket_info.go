package worker

type LeftTicketInfo struct {
	SecretStr        string // 下单用的密钥
	CanOrder         bool   // 是否接受预订
	TrainCode        string // 车次
	TrainNumber      string // 列车代号，订票排队用
	LeftTicketStr    string // 余票密钥串，订票排队用
	CandidateFlag    bool   // 是否可以候补
	Start            string // 始发站
	End              string // 终到站
	From             string // 出发站
	To               string // 到达站
	StartTime        string // 出发时间
	ArriveTime       string // 到达时间
	Duration         string // 历时
	ShangWuZuo       string // 商务座
	TeDengZuo        string // 特等座
	YiDengZuo        string // 一等座
	ErDengZuo        string // 二等座/二等包座
	GaoJiRuanWo      string // 高级软卧
	RuanWo           string // 软卧/一等卧
	DongWo           string // 动卧
	YingWo           string // 硬卧/二等卧
	RuanZuo          string // 软座
	YingZuo          string // 硬座
	WuZuo            string // 无座
	QiTa             string // 其他
	LeftTicketsCount []int  // 各类型座位的剩余票数
}
