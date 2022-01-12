package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/config"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"go.uber.org/zap"
)

type LoginConfResult struct {
	IsUAMLogin        bool `json:"is_uam_login"` // 统一认证登录（图片验证码）
	IsLoginPassCode   bool `json:"is_login_passCode"`
	IsSweepLogin      bool `json:"is_sweep_login"`
	IsMessagePassCode bool `json:"is_message_passCode"`
	IsLogin           bool `json:"is_login"` // 是否已登录

	LoginUrl string `json:"login_url"` // 登录 URL
	QueryUrl string `json:"queryUrl"`  // 查询余票 URL

	Now int64 `json:"now"` // 用于同步服务器时间

	StudentControl int `json:"stu_control"`   // 学生票提前开售天数
	OtherControl   int `json:"other_control"` // 其他票提前开售天数
}

// loginConf 获取登录设置
func loginConf(jar *cookiejar.Jar) (info *LoginConfResult, err error) {
	const (
		url0    = "https://%s/otn/login/conf"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	buf := bytes.NewBuffer([]byte("{}"))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	var (
		body       []byte
		statusCode int
	)
	if body, statusCode, err = httpcli.DoHttp(req, jar); err != nil {
		logger.Error("获取登录设置错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取登录设置失败", zap.Int("statusCode", statusCode), zap.ByteString("res", body))

		return nil, errors.New("get need captcha failure")
	}

	logger.Debug("获取登录设置", zap.ByteString("body", body))

	type LoginConfData struct {
		IsUAMLogin        string `json:"is_uam_login"` // 统一认证登录（图片验证码）
		IsLoginPassCode   string `json:"is_login_passCode"`
		IsSweepLogin      string `json:"is_sweep_login"`
		IsMessagePassCode string `json:"is_message_passCode"`
		IsLogin           string `json:"is_login"` // 是否已登录

		LoginUrl string `json:"login_url"` // 登录 URL
		QueryUrl string `json:"queryUrl"`  // 查询余票 URL

		Now int64 `json:"now"` // 用于同步服务器时间

		StudentControl int `json:"stu_control"`   // 学生票提前开售天数
		OtherControl   int `json:"other_control"` // 其他票提前开售天数
	}

	type LoginConfResponse struct {
		Status   bool          `json:"status"`
		Messages []string      `json:"messsages,omitempty"`
		Data     LoginConfData `json:"data"`
	}
	response := LoginConfResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析获取登录设置返回错误", zap.ByteString("res", body), zap.Error(err))

		return nil, err
	}

	if !response.Status {
		logger.Error("获取登录设置失败", zap.Int("statusCode", statusCode), zap.ByteString("res", body))

		return nil, errors.New(strings.Join(response.Messages, ""))
	}

	info = &LoginConfResult{
		IsUAMLogin:        response.Data.IsUAMLogin == "Y",
		IsLoginPassCode:   response.Data.IsLoginPassCode == "Y",
		IsSweepLogin:      response.Data.IsSweepLogin == "Y",
		IsLogin:           response.Data.IsLogin == "Y",
		IsMessagePassCode: response.Data.IsMessagePassCode == "" || response.Data.IsMessagePassCode == "Y",
		LoginUrl:          response.Data.LoginUrl,
		QueryUrl:          response.Data.QueryUrl,
		Now:               response.Data.Now,
		StudentControl:    response.Data.StudentControl,
		OtherControl:      response.Data.OtherControl,
	}

	// 预售天数
	config.Cfg.StudentPresellDays = response.Data.StudentControl
	config.Cfg.OtherPresellDays = response.Data.OtherControl

	return
}
