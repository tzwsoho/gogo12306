package candidate

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gogo12306/cdn"
	"gogo12306/httpcli"
	"gogo12306/logger"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

type CheckFaceRequest struct {
	SecretStr string
}

// CheckFace 获取人脸识别核验状态（原 12306 API 为 afterNate/chechFace，拼写错误？）
// https://kyfw.12306.cn/otn/resources/merged/queryLeftTicket_end_js.js 关键词: an 函数
func CheckFace(jar *cookiejar.Jar, info *CheckFaceRequest) (err error) {
	const (
		url0    = "https://%s/otn/afterNate/chechFace"
		referer = "https://kyfw.12306.cn/otn/leftTicket/init"
	)

	payload := &url.Values{}
	payload.Add("secretList", info.SecretStr)
	payload.Add("_json_att", "")

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)
	req.Header.Add("If-Modified-Since", "0")
	req.Header.Add("Cache-Control", "no-cache")

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("获取人脸识别核验状态错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取人脸识别核验状态失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("check face failure")
	}

	logger.Debug("获取人脸识别核验状态", zap.ByteString("body", body))

	type CheckFaceData struct {
		LoginFlag     bool   `json:"login_flag"`
		FaceFlag      bool   `json:"face_flag"`
		FaceCheckCode string `json:"face_check_code"`
		IsShowQRCode  bool   `json:"is_show_qrcode"`
	}

	type CheckFaceResponse struct {
		Status   bool          `json:"status"`
		Messages []string      `json:"messages"`
		Data     CheckFaceData `json:"data"`
	}
	response := CheckFaceResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析获取人脸识别核验状态返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	if !response.Status {
		logger.Error("获取人脸识别核验状态失败", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	} else if !response.Data.LoginFlag {
		logger.Error("获取人脸识别核验状态结果: 未登录不能候补", zap.Strings("错误消息", response.Messages))

		return errors.New(strings.Join(response.Messages, ""))
	} else if !response.Data.FaceFlag { // 未通过人脸识别，以下进入 bE 函数的逻辑
		logger.Info("获取人脸识别核验状态结果: 未通过人脸识别")

		switch response.Data.FaceCheckCode {
		case "01", "11":
			logger.Warn("证件信息正在审核中，请您耐心等待，审核通过后可继续完成候补操作。")

		case "03", "13":
			logger.Warn("证件信息审核失败，请检查所填写的身份信息内容与原证件是否一致。")

		case "04", "14":
			if response.Data.IsShowQRCode {
				// 下载人脸识别流程二维码并让用户扫码完成核验
				if err = GetCheckFaceQRCode(jar, &GetCheckFaceQRCodeRequest{
					AuthType:    "queueOrder",
					RiskChannel: "HB",
					CheckUrl:    "/afterNateQRCode/getClickScanStatus",
				}); err != nil {
					return
				}
			} else {
				logger.Warn("通过人证一致性核验的用户及激活的“铁路畅行”会员可以提交候补需求，请您按照操作说明在铁路12306app上完成人证核验")
			}

		case "02", "12": // 原 bE 函数就是留空不处理的
			return

		default:
			logger.Warn("获取人脸识别核验状态: 系统忙，请稍后再试!!!")
		}

		return errors.New("check face not pass")
	}

	return
}

type GetCheckFaceQRCodeRequest struct {
	AuthType    string
	RiskChannel string
	CheckUrl    string
}

// GetCheckFaceQRCode 获取人脸识别流程二维码
// 12306 源码里的 get_QRcodeAjax 函数
func GetCheckFaceQRCode(jar *cookiejar.Jar, info *GetCheckFaceQRCodeRequest) (err error) {
	const referer = ""
	url0 := "https://%s/otn" + info.CheckUrl

	payload := &url.Values{}
	payload.Add("appid", "otn")
	payload.Add("authType", info.AuthType)
	payload.Add("riskChannel", info.RiskChannel)

	buf := bytes.NewBuffer([]byte(payload.Encode()))
	req, _ := http.NewRequest("POST", fmt.Sprintf(url0, cdn.GetCDN()), buf)
	req.Header.Set("Referer", referer)
	httpcli.DefaultHeaders(req)

	var (
		body       []byte
		statusCode int
	)
	body, statusCode, err = httpcli.DoHttp(req, jar)
	if err != nil {
		logger.Error("获取人脸识别流程二维码错误", zap.Error(err))

		return
	} else if statusCode != http.StatusOK {
		logger.Error("获取人脸识别流程二维码失败", zap.Int("statusCode", statusCode), zap.ByteString("body", body))

		return errors.New("get check face qrcode failure")
	}

	logger.Debug("获取人脸识别流程二维码", zap.ByteString("body", body))

	type GetCheckFaceQRCodeResponse struct {
		ResultCode int    `json:"result_code"`
		UUID       string `json:"uuid,omitempty"`
		Image      string `json:"image,omitempty"`
	}
	response := GetCheckFaceQRCodeResponse{}
	if err = json.Unmarshal(body, &response); err != nil {
		logger.Error("解析获取人脸识别流程二维码返回错误", zap.ByteString("body", body), zap.Error(err))

		return
	}

	var decImage []byte
	if decImage, err = base64.StdEncoding.DecodeString(response.Image); err != nil {
		logger.Error("Base64 解码二维码图片错误", zap.String("image", response.Image), zap.Error(err))

		return
	}

	if err = ioutil.WriteFile("check_face_qrcode.png", decImage, 0666); err != nil {
		logger.Error("写入二维码图片错误", zap.String("image", response.Image), zap.Error(err))

		return
	}

	logger.Warn("人脸识别流程二维码已下载到 check_face_qrcode.png，请使用 12306 APP 扫描并完成人脸识别，然后再尝试候补车票")
	return
}
