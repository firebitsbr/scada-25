package main

import (
	"global"
	"scada/control/iec101"
	"scada/ctable"
	"scada/das/das_iec101/iec101"
	_ "scada/das/das_iec101/sync"

	"scada/send/opapi2op"
	"scada/serialconn"
	"time"
	"web"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

func main() {
	const CONF_FILE = "app.conf"
	const CONF_DRIVER = "table"
	reset()

	conf, err := config.NewConfig(CONF_DRIVER, ctable.GetAppName())
	if err != nil {
		return
	}

	setLogLevel(conf.String("app_logs_level"))
	logs.Info("==============================================")
	logs.Info("IEC101 传输原因,公共地址的长度是1,信息体地址长度是2.\n")

	//开启服务运行标识
	global.Start()

	//创建数据发送线程
	err = opapi2op.Run(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	//创建数据取数连接池
	pool, err := serialconn.New(conf.String("source_address"), 1)
	if err != nil {
		logs.Alert(err)
		return
	}

	//创建服务器控制服务
	ctrl, err := ctrl_iec101.New(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	go web.Run(
		conf.DefaultString("web_listen_address", ":80"),
		conf.DefaultString("web_path", "www"))

	//创建数据采集服务
	das, err := iec101.New(conf)
	if err != nil {
		logs.Error(err)
		return
	}

	for global.IsRunning() {
		conn, err := pool.Get()
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}

		err = das.Work(conn, ctrl)
		if err != nil {
			logs.Error(err)
			conn.Close()
			time.Sleep(time.Second * 2)
			continue
		}

		err = pool.Put(conn)
		if err != nil {
			conn.Close()
		}
	}
}
