package main

import (
	"flag"
	"gogo12306/cdn"
	"gogo12306/common"
	"gogo12306/config"
	"gogo12306/cookie"
	"gogo12306/logger"
	"gogo12306/login"
	"gogo12306/ticket"
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
			logger.Info("筛选可用 CDN", zap.Bool("cdn", *isCDN))

			cdn.FilterCDN(config.Cfg.CDN.CDNPath, config.Cfg.CDN.GoodCDNPath)
			return

		case "-g": // 开始抢票
			logger.Info("开始抢票", zap.Bool("grab", *isGrab))

			common.CheckOperationPeriod()

			var err error
			if err = cdn.LoadCDN(config.Cfg.CDN.GoodCDNPath); err != nil {
				return
			}

			///////////////////////////////////////////////////////////////////////////////////////////////////////////
			// 站点信息
			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			if err = ticket.InitStations(); err != nil {
				return
			}

			if err = ticket.InitSaleTime(); err != nil {
				return
			}

			if err = ticket.InitLeftTickerURL(); err != nil {
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

			if err = cookie.SetCookie(jar,
				config.Cfg.Login.GetCookieMethod,
				config.Cfg.Login.ChromeDriverPath,
				config.Cfg.Login.RailExpiration,
				config.Cfg.Login.RailDeviceID,
			); err != nil {
				return
			}

			// 先登录，好处时后面购票时不用再花时间登录，抢到票的几率增大
			// 但也有可能遇到当余票足够准备下单时，系统已自动退出登录，还是需要重新登录
			if config.Cfg.Login.Username != "" && config.Cfg.Login.Password != "" {
				if err = login.Login(jar); err != nil {
					return
				}

				login.CheckLoginTimer(jar)
			}

			///////////////////////////////////////////////////////////////////////////////////////////////////////////
			// 刷票任务
			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			for _, taskCfg := range config.Cfg.Tasks {
				var task *worker.Task
				if task, err = ticket.ParseTask(&taskCfg); err != nil || task == nil {
					logger.Error("转换任务配置出现错误", zap.Any("任务配置", taskCfg), zap.Error(err))
					return
				}

				worker.DoTask(jar, task)
			}

			///////////////////////////////////////////////////////////////////////////////////////////////////////////

			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
			<-c

			return
		}
	}

	flag.Usage()
}
