package main

import (
	"flag"
	"gogo12306/captcha"
	"gogo12306/cdn"
	"gogo12306/config"
	"gogo12306/logger"
	"gogo12306/notify/serverchan"
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
		config.Cfg.IsDevEnv,
		config.Cfg.LogFilepath,
		config.Cfg.LogLevel,
		config.Cfg.LogSplitMBSize,
		config.Cfg.LogKeepDays,
	)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-c": // 筛选可用 CDN
			logger.Debug("筛选可用 CDN", zap.Bool("cdn", *isCDN))

			cdn.FilterCDN(config.Cfg.CDNPath, config.Cfg.GoodCDNPath)
			return

		case "-t": // 测试消息发送
			logger.Debug("测试消息发送", zap.Bool("isTest", *isTest))

			serverchan.Notify("测试消息发送")
			return

		case "-o": // 测试验证码自动识别
			logger.Debug("测试验证码自动识别", zap.Bool("isOCRCaptcha", *isOCRCaptcha))

			var (
				err                      error
				base64Img, captchaResult string
			)
			if base64Img, err = captcha.GetCaptcha(); err != nil {
				return
			}

			t0 := time.Now()
			if captchaResult, err = captcha.GetCaptchaResult(config.Cfg.OCRUrl, base64Img); err != nil {
				return
			}

			logger.Debug("校验码结果", zap.String("ret", captchaResult), zap.Duration("耗时", time.Now().Sub(t0)))
			return

		case "-g": // 开始抢票
			logger.Debug("开始抢票", zap.Bool("grab", *isGrab))

			var err error
			if err = cdn.LoadCDN(config.Cfg.GoodCDNPath); err != nil {
				return
			}

			return

		default:

		}
	}

	flag.Usage()
}
