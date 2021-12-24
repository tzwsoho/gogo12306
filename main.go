package main

import (
	"flag"
	"gogo12306/captcha"
	"gogo12306/cdn"
	"gogo12306/config"
	"gogo12306/logger"
	"gogo12306/notifier"
	"gogo12306/station"
	"gogo12306/worker"
	"math/rand"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	isCDN := flag.Bool("c", false, "筛选可用 CDN")
	isMessage := flag.Bool("m", false, "测试消息发送")
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
		rand.Seed(time.Now().UnixNano())

		switch os.Args[1] {
		case "-c": // 筛选可用 CDN
			logger.Debug("筛选可用 CDN", zap.Bool("cdn", *isCDN))

			cdn.FilterCDN(config.Cfg.CDN.CDNPath, config.Cfg.CDN.GoodCDNPath)
			return

		case "-m": // 测试消息发送
			logger.Debug("测试消息发送", zap.Bool("isMessage", *isMessage))

			notifier.Broadcast("测试消息发送")
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

			// 获取校验码图片 BASE64
			if base64Img, err = captcha.GetCaptcha(jar); err != nil {
				return
			}

			// 自动识别校验码
			t0 := time.Now()
			if err = captcha.GetCaptchaResult(jar, base64Img, &captchaResult); err != nil {
				return
			}
			deltaT := time.Now().Sub(t0)

			// 将识别结果转化为坐标点
			answer := captcha.ConvertCaptchaResult(&captchaResult)

			// 验证校验码结果
			if pass, err = captcha.VerifyCaptcha(jar, answer); err != nil {
				return
			}

			logger.Debug("校验码验证结果",
				zap.Any("校验码 OCR 结果", captchaResult.Result),
				zap.String("转化后坐标点", answer),
				zap.Bool("校验码验证是否通过", pass),
				zap.Duration("自动识别校验码耗时", deltaT),
			)
			return

		case "-g": // 开始抢票
			logger.Debug("开始抢票", zap.Bool("grab", *isGrab))

			var err error
			if err = cdn.LoadCDN(config.Cfg.CDN.GoodCDNPath); err != nil {
				return
			}

			///////////////////////////////////////////////////////////////////////////////////////////////////////////
			// 站点信息
			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			if err = station.InitStations(); err != nil {
				return
			}

			if err = station.InitSaleTime(); err != nil {
				return
			}

			///////////////////////////////////////////////////////////////////////////////////////////////////////////
			// 登录
			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			var jar *cookiejar.Jar
			if jar, err = cookiejar.New(nil); err != nil {
				logger.Error("创建 Jar 错误", zap.Error(err))
				return
			}

			var needCaptcha bool
			if needCaptcha, err = captcha.NeedCaptcha(jar); err != nil {
				return
			}
			_ = needCaptcha

			// if config.Cfg.Login.Username != "" && config.Cfg.Login.Password != "" {
			// 	if err = login.Login(jar, needCaptcha); err != nil {
			// 		return
			// 	}

			// 	// 定时检查登录状态
			// 	login.CheckLoginTimer(jar)
			// }

			///////////////////////////////////////////////////////////////////////////////////////////////////////////
			// 刷票任务
			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			if len(config.Cfg.Tasks) > 0 {
				for _, t := range config.Cfg.Tasks {
					task := &worker.Task{}
					if err = task.Parse(&t); err != nil {
						logger.Error("转换任务配置出现错误", zap.Any("任务配置", t), zap.Error(err))
						return
					}

					worker.DoTask(task)
				}
			}

			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
			<-c
		}
	}

	flag.Usage()
}
