package syncdb

import (
	"fmt"
	"os"
	"scada/ctable"
	"scada/points"

	"github.com/astaxie/beego/logs"
)

//关系库同步
func init() {
	for _, s := range os.Args {
		if s == "-sync" {
			app := points.GetAppName()
			//需要同步点表
			rows := MakeRows()
			err := ctable.Sync(rows, app)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	//如果没有参数时, 查询数据库是否有配置文件,如果数据库没有配置文件时生成默认配置
	app := points.GetAppName()
	n, err := ctable.ConfCount(app)
	if err != nil {
		logs.Error("Query configure ", err)
		return
	}

	if n == 0 {
		rows := MakeRows()
		err = ctable.Sync(rows, app)
		if err != nil {
			logs.Error("Query configure ", err)
		}
	}

}

func MakeRows() (rows []*ctable.ConfTable) {
	rows = make([]*ctable.ConfTable, 0)
	row := new(ctable.ConfTable)
	row.Name = "程序日志级别"
	row.Key = "app_logs_level"
	row.Value = "Info"
	row.Desc = "日志级别有:Debug,Info,Notice,Warning,Error,Alert,Emergency"
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "PPI 从站地址"
	row.Key = "source_address"
	row.Value = "com1,9600,n,8,1"
	row.Desc = "Modbus TCP 的服务器地址,端口一般默认为502"
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "PPI 数据的读取间隔"
	row.Key = "source_interval"
	row.Value = "1000"
	row.Desc = "循环读取一次的间隔时间,单位毫秒"
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "发送数据的服务器地址"
	row.Key = "destination_address"
	row.Value = "127.0.0.1:8200"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "发送数据服务器的用户名"
	row.Key = "destination_user_name"
	row.Value = "sis"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "发送数据服务器的密码"
	row.Key = "destination_user_password"
	row.Value = "openplant"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "发送数据服务器的选项"
	row.Key = "destination_option"
	row.Value = "0"
	row.Desc = "0: 写入数据到服务器没有隔离器, 2:写入数据到服务器有隔离器"
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "PPI目标地址"
	row.Key = "ppi.da"
	row.Value = "02"
	row.Desc = "指PLC在PPI上地址，一台PLC时，一般为02，多台PLC时，则各有各的地址"
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "PPI源地址"
	row.Key = "ppi.da"
	row.Value = "00"
	row.Desc = "指计算机在PPI上地址，一般为00"
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "提供WEB服务的地址"
	row.Key = "web_listen_address"
	row.Value = ":8080"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ctable.ConfTable)
	row.Name = "提供WEB服务的路径"
	row.Key = "web_path"
	row.Value = "../../www"
	row.Desc = "WEB相关文件所在的路径"
	rows = append(rows, row)
	return
}
