package main

import (
	"global"
	"scada/control/iec101"
	"scada/ctable"
	"scada/das/das_iec104/iec104"
	_ "scada/das/das_iec104/sync"
	"scada/send/opapi2op"
	"scada/tcpconn"
	"time"
	"web"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

func main() {
	//const CONF_DRIVER = "ini"
	const CONF_DRIVER = "table"
	reset()

	conf, err := config.NewConfig(CONF_DRIVER, ctable.GetAppName())
	if err != nil {
		return
	}

	setLogLevel(conf.String("app_logs_level"))
	logs.Info("==============================================")
	logs.Info("IEC104 默认的端口号是2404,传输原因,公共地址的长度是2,信息体地址长度是3.\n")

	//开启服务运行标识
	global.Start()

	//创建数据发送线程
	err = opapi2op.Run(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	//创建数据取数连接池
	pool, err := tcpconn.New(conf.String("source_address"), 1)
	if err != nil {
		logs.Alert(err)
		return
	}

	//创建服务器控制服务,IEC104 使用的是101的控制方法
	ctrl, err := ctrl_iec101.New(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	go web.Run(
		conf.DefaultString("web_listen_address", ":80"),
		conf.DefaultString("web_path", "www"))

	//创建数据采集服务
	das, err := iec104.New(conf)
	if err != nil {
		return
	}

	for global.IsRunning() {
		das.Check()
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
