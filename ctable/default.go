package ctable

import (
	"log"
	"os"
)

func Save2file(rows []*ConfTable) {
	f, err := os.Create("config.txt")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	//f.WriteString("Id,Desc,Scope,Tips,Default,TextType,TextData,Validate\n")
	f.WriteString("Id\tDesc\tScope\tTips\tDefault\tTextType\tTextData\tValidate\n")
	for _, row := range rows {
		f.WriteString(row.Key + "\t")
		f.WriteString(row.Name + "\t")
		f.WriteString(row.Scope + "\t")
		f.WriteString(row.Desc + "\t")
		f.WriteString(row.Value + "\t")
		f.WriteString(row.TextType + "\t")
		f.WriteString(row.TextData + "\t\n")
	}
}

//添加默认的测点字段属性
func AddDefaultPointRows() []*ConfTable {
	rows := make([]*ConfTable, 0)
	row := AddRow("ID", "Id", "", "测点ID", "addPoint", "", "")
	rows = append(rows, row)
	row = AddRow("测点名称", "Pn", "", "测点名称", "addPoint", "", "")
	rows = append(rows, row)
	row = AddRow("数据类型", "Rt", "", "测点类型", "addPoint", "select", "AX|DX|I2|I4|R8")
	rows = append(rows, row)
	row = AddRow("地址信息", "Ad", "", "地址信息,多个属性用;分割", "addPoint", "", "")
	rows = append(rows, row)
	row = AddRow("描述", "Ed", "", "描述", "addPoint", "", "")
	rows = append(rows, row)
	row = AddRow("系数", "Fk", "1", "实时值*系数", "addPoint", "", "")
	rows = append(rows, row)
	row = AddRow("基数", "Fb", "0", "实时值+基数", "addPoint", "", "")
	rows = append(rows, row)
	return rows
}

func AddDefaultWebRows() []*ConfTable {
	rows := make([]*ConfTable, 0)
	row := AddRow("提供WEB服务的地址", "web_listen_address", ":8080", "",
		"addServer", "", "")
	rows = append(rows, row)
	row = AddRow("提供WEB服务的路径", "web_path", "../www", "WEB相关文件所在的路径",
		"addServer", "", "")
	rows = append(rows, row)
	return rows
}

//添加默认的通用配置
func AddDefaultRows() []*ConfTable {
	rows := make([]*ConfTable, 0)
	row := AddRow("程序日志级别", "app_logs_level", "Info", "程序记录的日志级别",
		"addServer", "select", "Debug|Info|Notice|Warning|Error|Alert|Emergency")
	rows = append(rows, row)
	return rows
}

//添加默认的控制配置属性
func AddDefaultControlRows() []*ConfTable {
	rows := make([]*ConfTable, 0)

	row := AddRow("测点属性", "Event", "0",
		"0:读 1:写 2:读/写", "addPoint", "select", "0|1|2")
	rows = append(rows, row)

	row = AddRow("是否开启服务器控制功能", "control_enable", "false",
		"false: 不开启服务器控制 true: 开启服务器控制", "addServer", "select", "false|true")
	rows = append(rows, row)

	row = AddRow("下发控制命令的服务器", "control_address", "127.0.0.1:8200",
		"控制到采集之间不能有隔离器", "addServer", "", "")
	rows = append(rows, row)

	row = AddRow("下发控制命令的服务器用户名", "control_user_name", "sis",
		"需要有控制权限的用户名", "addServer", "", "")
	rows = append(rows, row)

	row = AddRow("下发控制命令的服务器密码", "control_user_password", "openplant",
		"", "addServer", "password", "")
	rows = append(rows, row)

	return rows
}

//添加默认的发送属性
func AddDefaultSendRows() []*ConfTable {
	rows := make([]*ConfTable, 0)

	row := AddRow("发送数据的服务器地址", "destination_address", "127.0.0.1:8200",
		"", "addServer", "", "")
	rows = append(rows, row)

	row = AddRow("发送数据服务器的用户名", "destination_user_name", "sis",
		"需要有写入权限的用户名", "addServer", "", "")
	rows = append(rows, row)

	row = AddRow("发送数据服务器的密码", "destination_user_password", "openplant",
		"", "addServer", "password", "")
	rows = append(rows, row)

	row = AddRow("发送数据服务器的选项", "destination_option", "0",
		"0: 写入数据到服务器没有隔离器, 2:写入数据到服务器有隔离器", "addServer", "select", "0|2")
	rows = append(rows, row)
	return rows
}
