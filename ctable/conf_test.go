package ctable

import (
	"log"
	"testing"
)

func TestConf(t *testing.T) {
	rows := make([]*ConfTable, 0)
	row := new(ConfTable)
	row.Name = "程序日志级别"
	row.Key = "app.logs.level"
	row.Value = "Info"
	row.Desc = "日志级别有:Debug,Info,Notice,Warning,Error,Alert,Emergency"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "Modbus 从站地址"
	row.Key = "source.address"
	row.Value = "127.0.0.1:502"
	row.Desc = "Modbus TCP 的服务器地址,端口一般默认为502"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "Modbus 数据的读取间隔"
	row.Key = "source.interval"
	row.Value = "1000"
	row.Desc = "循环读取一次的间隔时间,单位毫秒"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "发送数据的服务器地址"
	row.Key = "destination.address"
	row.Value = "127.0.0.1:8200"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "发送数据服务器的用户名"
	row.Key = "destination.user.name"
	row.Value = "sis"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "发送数据服务器的密码"
	row.Key = "destination.user.password"
	row.Value = "openplant"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "发送数据服务器的选项"
	row.Key = "destination.option"
	row.Value = "0"
	row.Desc = "0: 写入数据到服务器没有隔离器, 2:写入数据到服务器有隔离器"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "是否开启服务器控制功能"
	row.Key = "control.enable"
	row.Value = "false"
	row.Desc = "false: 不开启服务器控制 true: 开启服务器控制"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "下发控制命令的服务器"
	row.Key = "control.address"
	row.Value = "127.0.0.1:8200"
	row.Desc = "控制到采集之间不能有隔离器,这个地址通常与发送数据库地址相同"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "下发控制命令的服务器用户名"
	row.Key = "control.user.name"
	row.Value = "sis"
	row.Desc = "需要有控制权限的用户名"
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "下发控制命令的服务器密码"
	row.Key = "control.user.password"
	row.Value = "openplant"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "提供WEB服务的地址"
	row.Key = "web.listen.address"
	row.Value = ":80"
	row.Desc = ""
	rows = append(rows, row)

	row = new(ConfTable)
	row.Name = "提供WEB服务的路径"
	row.Key = "web.path"
	row.Value = "../../www"
	row.Desc = "WEB相关文件所在的路径"
	rows = append(rows, row)
	err := Sync(rows)
	if err != nil {
		log.Println(err)
	}
}
