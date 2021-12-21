package main

import (
	"flag"
	"gogo12306/captcha"
	"gogo12306/cdn"
	"gogo12306/config"
	"gogo12306/logger"
	"gogo12306/login"
	"gogo12306/notify/serverchan"
	"net/http/cookiejar"
	"os"
	"time"

	"go.uber.org/zap"
)

func main() {
	isCDN := flag.Bool("c", false, "筛选可用 CDN")
	isTest := flag.Bool("t", false, "测试消息发送")
	isOCRCaptcha := flag.Bool("o", false, "测试验证码自动识别")
	isGrab := flag.Bool("g", false, "开始抢票")
	flag.Parse()

	config.Init("config.json")

	logger.Init(
		config.Cfg.Logger.IsDevelop,
		config.Cfg.Logger.LogFilepath,
		config.Cfg.Logger.LogLevel,
		config.Cfg.Logger.LogSplitMBSize,
		config.Cfg.Logger.LogKeepDays,
	)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-c": // 筛选可用 CDN
			logger.Debug("筛选可用 CDN", zap.Bool("cdn", *isCDN))

			cdn.FilterCDN(config.Cfg.CDN.CDNPath, config.Cfg.CDN.GoodCDNPath)
			return

		case "-t": // 测试消息发送
			logger.Debug("测试消息发送 TODO", zap.Bool("isTest", *isTest))

			serverchan.Notify("测试消息发送")
			return

		case "-o": // 测试验证码自动识别
			logger.Debug("测试验证码自动识别", zap.Bool("isOCRCaptcha", *isOCRCaptcha))

			var (
				err           error
				jar           *cookiejar.Jar
				base64Img     string
				captchaResult captcha.CaptchaResult
				pass          bool
			)
			if jar, err = cookiejar.New(nil); err != nil {
				logger.Error("创建 Jar 错误", zap.Error(err))
				return
			}

			if base64Img, err = captcha.GetCaptcha(jar); err != nil {
				return
			}

			t0 := time.Now()
			if err = captcha.GetCaptchaResult(jar, config.Cfg.OCR.OCRUrl, base64Img, &captchaResult); err != nil {
				return
			}

			answer := captcha.ConvertCaptchaResult(&captchaResult)

			if pass, err = captcha.CheckCaptcha(jar, answer); err != nil {
				return
			}

			logger.Debug("校验码结果",
				zap.Any("校验码 OCR 结果", captchaResult.Result),
				zap.String("转化后结果", answer),
				zap.Bool("校验码验证是否通过", pass),
				zap.Duration("耗时", time.Now().Sub(t0)),
			)
			return

		case "-g": // 开始抢票
			logger.Debug("开始抢票 TODO", zap.Bool("grab", *isGrab))

			var err error
			if err = cdn.LoadCDN(config.Cfg.CDN.GoodCDNPath); err != nil {
				return
			}

			if err = login.Login(); err != nil {
				return
			}

			return

		default:

		}
	}

	flag.Usage()
}
