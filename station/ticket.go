package station

import (
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"regexp"

	"go.uber.org/zap"
)

var leftTicketURL string
var loginIsDisable bool
var checkUserMDID string

func InitLeftTickerURL() (err error) {
	const (
		url     = "https://%s/otn/leftTicket/init"
		referer = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req, _ := http.NewRequest("GET", fmt.Sprintf(url, cdn.GetCDN()), nil)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, nil)
	if err != nil {
		logger.Error("获取余票查询 URL 错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取余票查询 URL 失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get left ticket url failure")
	}

	var re1, re2, re3 *regexp.Regexp
	if re1, err = regexp.Compile("var CLeftTicketUrl\\s*=\\s*(?:'|\")([^'\"]+?)(?:'|\")"); err != nil {
		logger.Error("生成正则表达式 1 失败", zap.Error(err))

		return
	}

	if re2, err = regexp.Compile("var login_isDisable\\s*=\\s*(?:'|\")([^'\"]+?)(?:'|\")"); err != nil {
		logger.Error("生成正则表达式 2 失败", zap.Error(err))

		return
	}

	if re3, err = regexp.Compile("var checkusermdId\\s*=\\s*(?:'|\")([^'\"]+?)(?:'|\")"); err != nil {
		logger.Error("生成正则表达式 3 失败", zap.Error(err))

		return
	}

	body1 := re1.FindSubmatch(body)
	if body1 == nil || len(body1) != 2 {
		logger.Error("匹配正则表达式 1 失败", zap.ByteString("body", body), zap.String("re", re1.String()))

		return errors.New("regexp 1 match failure")
	}

	body2 := re2.FindSubmatch(body)
	if body2 == nil || len(body2) != 2 {
		logger.Error("匹配正则表达式 2 失败", zap.ByteString("body", body), zap.String("re", re2.String()))

		return errors.New("regexp 2 match failure")
	}

	body3 := re3.FindSubmatch(body)

	leftTicketURL = string(body1[1])
	loginIsDisable = (string(body2[1]) == "Y")

	if body3 != nil && len(body3) == 2 {
		checkUserMDID = string(body3[1])
	}

	return
}
