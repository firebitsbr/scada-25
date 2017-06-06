package main

import (
	"global"
	"scada/ctable"
	"scada/das/ppi/ppi"
	"scada/points"
	"scada/send/opapi2op"
	"scada/tcpconn"
	"time"
	"web"

	_ "scada/das/ppi/sync"

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
	err = opapi2op.Run(conf)
	if err != nil {
		logs.Alert(err)
		return
	}

	go web.Run(
		conf.DefaultString("web_listen_address", ":80"),
		conf.DefaultString("web_path", "www"))

	//创建数据取数连接池
	pool, err := tcpconn.New(conf.String("source_address"), 1)
	if err != nil {
		logs.Alert(err)
		return
	}

	//创建数据采集服务
	das, err := ppi.New(points.GetAppName())
	if err != nil {
		return
	}

	for global.IsRunning() {
		conn, err := pool.Get()
		if err != nil {
			time.Sleep(time.Second * 2)
			continue
		}
		err = das.Work(conn, nil)
		if err != nil {
			conn.Close()
			time.Sleep(time.Second)
			continue
		}

		err = pool.Put(conn)
		if err != nil {
			conn.Close()
		}
		time.Sleep(time.Second)
	}
}
