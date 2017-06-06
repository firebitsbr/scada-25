package syncdb

import (
	"fmt"
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
			rows := MakeRows()
			err := ctable.Sync(rows, app, "example_conf")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
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

//数据转发 暂不支持控制功能
func MakeRows() (rows []*ctable.ConfTable) {
	row := ctable.AddRow("实时数据的读取间隔", "source_interval", "1000",
		"循环读取一次的间隔时间,单位毫秒", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("实时数据库源地址", "source_address", "127.0.0.1;8200",
		"", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("实时数据库源的用户名", "source_user_name", "sis",
		"", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("实时数据库源的密码", "source_user_password", "openplant",
		"", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	rs := ctable.AddDefaultRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	rs = ctable.AddDefaultSendRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	rs = ctable.AddDefaultWebRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	return
}
