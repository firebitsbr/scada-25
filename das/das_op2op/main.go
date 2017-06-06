package main

import (
	"global"
	"opapi3"
	"scada/ctable"
	"scada/das/das_op2op/op2op"
	"scada/points"
	"scada/send/udp2op"
	"web"

	_ "scada/das/das_op2op/sync"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

func main() {
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
	err = udp2op.Run(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	go web.Run(
		conf.DefaultString("web_listen_address", ":80"),
		conf.DefaultString("web_path", "www"))

	conn := opapi3.NewConnect(
		conf.String("source_address"),
		conf.String("source_user_name"),
		conf.String("source_user_password"),
		0)

	//创建数据采集服务
	//das, err := modbustcp.New(conf.String("das.points"))
	das, err := op2op.New(points.GetAppName())
	if err != nil {
		return
	}
	err = das.Work(conn, nil)
}
