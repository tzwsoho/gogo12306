# gogo12306
Go编写的12306抢票工具

# 使用方法
将 config.example.json 复制一份，更名为 config.json，然后打开后按注释修改里面的内容

## Usage of gogo12306:

-c    筛选延时在 300ms 内的可用 CDN

-g    开始抢票


# 编译：
go mod tidy

go build


# 运行：
## ①将config.example.json复制一份，命名为config.json，按里面的注释修改配置

## ②执行以下命令筛选延时300ms以内的CDN，加速刷票进度：

gogo12306 -c

## ③执行以下命令开始刷票：

gogo12306 -g

# 目前已完成的功能：
- [x] 自动识别验证码
- [x] 登录 12306
- [x] 使用加速 CDN 使刷票更快
- [x] 定时刷票
- [x] 自动下单
- [x] 候补订单
- [x] 抢票成功提醒（目前只支持 Server酱、WXPusher）

# 验证码识别依靠pjialin大佬的PY12306 助手 (pjialin.com)(https://py12306-helper.pjialin.com/)
# 感谢wenxianping大佬的开源项目(https://github.com/testerSunshine/12306)，GOGO12306有很多地方借鉴了这个项目