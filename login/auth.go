package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http"
	"net/http/cookiejar"

	"go.uber.org/zap"
)

func Auth(jar *cookiejar.Jar) (tk string, err error) {
	const (
		url1     = "https://%s/otn/resources/login.html"
		referer1 = "https://kyfw.12306.cn/otn/view/index.html"
	)
	req1, _ := http.NewRequest("GET", fmt.Sprintf(url1, cdn.GetCDN()), nil)
	req1.Header.Set("Referer", referer1)
	httpcli.DefaultHeaders(req1)

	var (
		body1       []byte
		statusCode1 int
	)
	if body1, statusCode1, err = httpcli.DoHttp(req1, jar); err != nil {
		logger.Error("授权登录错误", zap.Error(err))

		return "", err
	} else if statusCode1 != http.StatusOK {
		logger.Error("授权登录失败", zap.Int("statusCode", statusCode1), zap.ByteString("res", body1))

		return "", errors.New("login auth failure")
	}

	const (
		url2     = "https://%s/passport/web/auth/uamtk-static?appid=otn"
		referer2 = "https://kyfw.12306.cn/otn/resources/login.html"
	)
	req2, _ := http.NewRequest("GET", fmt.Sprintf(url2, cdn.GetCDN()), nil)
	req2.Header.Set("Referer", referer2)
	httpcli.DefaultHeaders(req2)

	var (
		body2       []byte
		statusCode2 int
	)
	if body2, statusCode2, err = httpcli.DoHttp(req2, jar); err != nil {
		logger.Error("授权错误", zap.Error(err))

		return "", err
	} else if statusCode2 != http.StatusOK {
		logger.Error("授权失败", zap.Int("statusCode", statusCode2), zap.ByteString("res", body2))

		return "", errors.New("auth failure")
	}

	type AuthResult struct {
		ResultCode    int    `json:"result_code"`
		ResultMessage string `json:"result_message"`
		Name          string `json:"name"`
		NewAppTK      string `json:"newapptk"`
	}
	result := AuthResult{}
	if err = json.Unmarshal(body2, &result); err != nil {
		logger.Error("解析授权返回信息错误", zap.ByteString("res", body2), zap.Error(err))

		return "", err
	}

	if result.ResultCode == 0 {
		logger.Info("授权成功",
			zap.String("用户", result.Name),
			zap.String("tk", result.NewAppTK))
	} else {
		logger.Error("授权失败，请使用官方 APP 登录并进行人脸校验再次尝试登录",
			zap.Int("code", result.ResultCode),
			zap.String("msg", result.ResultMessage))
	}

	return result.NewAppTK, nil
}
