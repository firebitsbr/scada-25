package syncdb

import (
	"os"
	"scada/ctable"

	"github.com/astaxie/beego/logs"
)

//关系库同步
func init() {
	app := ctable.GetAppName()
	for _, s := range os.Args {
		if s == "-sync" {
			//需要同步点表
			//需要同步点表
			ctable.Save2file(MakeRows())
			os.Exit(0)
		}
	}

	n, err := ctable.ConfCount(app)
	if err != nil {
		logs.Error("Query configure ", err)
		return
	}

	if n == 0 {
		rows := MakeRows()
		err = ctable.Sync(rows, app, "das_conf")
		if err != nil {
			logs.Error("Query configure ", err)
		}
	}
}

func MakeRows() (rows []*ctable.ConfTable) {
	rows = make([]*ctable.ConfTable, 0)
	rs := ctable.AddDefaultPointRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	row := ctable.AddRow("Modbus 从站地址", "source_address", "127.0.0.1:502",
		"Modbus TCP 的服务器地址,端口一般默认为502,多个地址用分号分割", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("Modbus 数据的读取间隔", "source_interval", "1000",
		"循环读取一次的间隔时间,单位毫秒", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	rs = ctable.AddDefaultRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	rs = ctable.AddDefaultSendRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	rs = ctable.AddDefaultControlRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	rs = ctable.AddDefaultWebRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	return
}
