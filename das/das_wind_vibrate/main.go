package main

import (
	"global"
	"log"
	"scada/ctable"
	"scada/das/das_wind_vibrate/wv"
	"scada/send/opapi2op"
	"time"
	"web"

	_ "scada/das/das_wind_forecast/sync"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	//logs.SetLevel(logs.LevelInfo)
	//const CONF_FILE = "app.conf"
	//const CONF_DRIVER = "ini"
	const CONF_DRIVER = "table"
	reset()

	conf, err := config.NewConfig(CONF_DRIVER, ctable.GetAppName())
	if err != nil {
		logs.Error("读取配置文件", err)
		return
	}
	setLogLevel(conf.String("app_logs_level"))

	//开启服务运行标识
	global.Start()

	//创建数据发送线程
	err = opapi2op.Run(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	go web.Run(
		conf.DefaultString("web_listen_address", ":80"),
		conf.DefaultString("web_path", "www"))

	//创建数据采集服务
	das, err := wv.New(conf)
	if err != nil {
		return
	}

	for global.IsRunning() {
		err = das.Work(nil, nil)
		time.Sleep(time.Second * 10)
	}
}
