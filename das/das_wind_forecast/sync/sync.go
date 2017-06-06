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

	row := ctable.AddRow("源文件存放的目录", "file_path", "data",
		"原始文件存放的目录", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("扫描文件间隔", "source_interval", "10",
		"每次扫描文件的周期,单位秒", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("备份文件存放的目录", "file_bak", "bak",
		"解析原始文件后,备份解析后的文件存放路径", "addServer", "", "")
	if row != nil {
		rows = append(rows, row)
	}

	row = ctable.AddRow("备份文件存放天数", "file_save_days", "30",
		"删除超过指定天数的备份文件", "addServer", "", "")
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

	rs = ctable.AddDefaultWebRows()
	if rs != nil {
		rows = append(rows, rs...)
	}

	return
}
