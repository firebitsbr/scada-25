package points

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"scada/ctable"
	"strconv"
	"strings"
	"time"
	"web"

	"github.com/astaxie/beego/logs"
)

//实时数据
type RtlValue struct {
	ID int    `json:"ID"`
	AV string `json:"AV"`
	AS string `json:"AS"`
	TM string `json:"TM"`
}

//点的质量统计
type PointQualityCount struct {
	Total   int
	Good    int
	Bad     int
	Timeout int
}

type Rows struct {
	Count int
	Rows  []*Row
}

type Row struct {
	UID int    //<全局ID
	PN  string //<点名
	ID  int32  //<ID
	RT  int32  //<OP类型
	AV  string
	AS  uint16  //<状态
	TM  int64   //<时间
	PT  int32   //<类型
	FB  float64 //<基数
	FK  float64 //<系数
	CP  int32   //<站号
	HW  int32   //<地址
	ET  string  //<扩展字段
	RK  string  //<备注
	ED  string  //<描述
	UP  int32   //<更新状态	=0 时表示要发送数据
}

func init() {
	web.HandleFunc("/todo", HttpCommand)
	web.HandleFunc("/do", HttpCommand)
	web.HandleFunc("/upload", HttpUpload)
	web.HandleFunc("/upload/points", HttpUploadPoints)
	web.HandleFunc("/upload/points/inform", HttpUploadPointsInform)
	web.HandleFunc("/upload/events", HttpUploadEvents)
	web.HandleFunc("/download/points.txt", HttpDownloadPoints)
	web.HandleFunc("/download/events.txt", HttpDownloadEvents)
	web.HandleFunc("/config.html", HttpConfig)
	web.HandleFunc("/reset", HttpReset)
}

type Sizer interface {
	Size() int64
}

type Columns struct {
	ID          int
	Name        int
	Type        int
	Desc        int
	Address     int
	Base        int
	Coefficient int
}

func ColumnsPosition(line string) (columns *Columns) {
	columns = new(Columns)
	columns.ID = -1
	columns.Name = -1
	columns.Type = -1
	columns.Desc = -1
	columns.Address = -1
	columns.Base = -1
	columns.Coefficient = -1
	line = strings.TrimSpace(line)
	line = strings.ToUpper(line)
	ss := strings.Split(line, ",")
	for i := 0; i < len(ss); i++ {
		ss[i] = strings.TrimSpace(ss[i])
	}
	for i := 0; i < len(ss); i++ {
		logs.Debug(i, ss[i])
		switch ss[i] {
		case "ID":
			columns.ID = i
		case "NAME":
			columns.Name = i
		case "TYPE":
			columns.Type = i
		case "DESC":
			columns.Desc = i
		case "ADDRESS":
			columns.Address = i
		case "BASE":
			columns.Base = i
		case "COEFFICIENT":
			columns.Coefficient = i
		}
	}
	return
}

func getType(s string) (t string) {
	t = "AX"
	s = strings.ToUpper(s)
	switch s {
	case "0", "AX":
		return "AX"
	case "1", "DX":
		return "DX"
	case "2", "I2":
		return "I2"
	case "3", "I4":
		return "I4"
	case "4", "R8":
		return "R8"
	}
	return
}

func GetAppName() string {
	return ctable.GetAppName()
}

//把一行数据格式化为sql语句
func FormatSqlFromLine(cols *Columns, line string, app string, isEvent bool) (sql string, err error) {
	ss := strings.Split(line, ",")
	s := bytes.NewBuffer(nil)
	strEvent := "','0',"
	if isEvent {
		strEvent = "','1',"
	}
	for i := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
	if len(ss) > 1 {
		s.WriteString("INSERT INTO point (sr,sn, event, uid, pn, rt, ed, ad, fb, fk) VALUES('','" +
			app + strEvent)
		if cols.ID >= 0 && cols.ID < len(ss){
			s.WriteString(ss[cols.ID])
		} else {
			s.WriteString("-1")
		}
		if cols.Name >= 0 && cols.Name < len(ss) {
			s.WriteString(",'" + ss[cols.Name] + "'")
		} else {
			s.WriteString(",''")
		}
		if cols.Type >= 0 && cols.Type < len(ss){
			s.WriteString(",'" + getType(ss[cols.Type]) + "'")
		} else {
			s.WriteString(",''")
		}
		if cols.Desc >= 0 && cols.Desc < len(ss) {
			s.WriteString(",'" + ss[cols.Desc] + "'")
		} else {
			s.WriteString(",''")
		}
		if cols.Address >= 0&& cols.Address < len(ss) {
			s.WriteString(",'" + ss[cols.Address] + "'")
		} else {
			s.WriteString(",''")
		}
		if cols.Base >= 0 && cols.Base < len(ss){
			s.WriteString("," + ss[cols.Base])
		} else {
			s.WriteString(",0")
		}
		if cols.Coefficient >= 0 && cols.Coefficient < len(ss){
			s.WriteString("," + ss[cols.Coefficient])
		} else {
			s.WriteString(",1")
		}
		s.WriteString(");")
		sql = s.String()
		//logs.Debug(sql)
	} else {
		err = errors.New("没有足够的列数")
	}
	return
}

func MakeConfigPage(w http.ResponseWriter) {
	//生成配置页面
	data := make([]*ctable.ConfTable, 0)
	app := GetAppName()
	if app == "" {
		w.Write([]byte("不能获取有效的服务名!"))
		return
	}
	name := ""
	key := ""
	value := ""
	desc := ""
	sqlString := "SELECT name, key, value, ed FROM das_conf WHERE ex_scope='addServer' and driver ='" + app + "' ORDER BY key ASC;"
	db := ctable.GetDB()
	if db == nil {
		w.Write([]byte("不能连接到数据库!"))
		return
	}

	rows, err := db.Query(sqlString)
	if err != nil {
		logs.Error("Export configure,", err)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &key, &value, &desc)
		if err != nil {
			logs.Info("SELECT FROM conf:", err)
			continue
		}
		row := new(ctable.ConfTable)
		row.Name = name
		row.Key = key
		row.Value = value
		row.Desc = desc
		data = append(data, row)
	}

	web.ServeHtml(w, "config.html", data)
}

func HttpReset(w http.ResponseWriter, r *http.Request) {
	err := ctable.ClearTable(ctable.GetAppName())
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	var cmd *exec.Cmd
	if len(os.Args) == 1 {
		cmd = exec.Command(os.Args[0])
	} else {
		cmd = exec.Command(os.Args[0], os.Args[1:]...)
	}
	err = cmd.Start()
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("OK"))
	}
	os.Exit(0)
}

//配置文件修改
func HttpConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm() //解析参数，默认是不会解析的
		//更新关系数据库
		app := GetAppName()
		if app == "" {
			w.Write([]byte("不能获取有效的服务名!"))
			return
		}
		conn := ctable.GetDB()
		if conn == nil {
			w.Write([]byte("连接到数据库失败!"))
			return
		}
		sqlString := ""
		tx, err := conn.Begin()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		for key, value := range r.Form {
			if len(value) == 0 {
				continue
			}
			sqlString = "UPDATE das_conf SET value=? WHERE driver=? AND key=?;"
			_, err = tx.Exec(sqlString, value[0], app, key)
			if err != nil {
				logs.Info(sqlString)
				w.Write([]byte(sqlString + err.Error()))
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("OK"))
	} else {
		//生成配置页面
		MakeConfigPage(w)
	}
}

//标识采集点表已经更新
func HttpUploadPointsInform(w http.ResponseWriter, r *http.Request) {
	pointsUpdate = true
	pointList.Reset()
	w.Write([]byte("OK"))
}

//上传点表
func HttpUploadPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		msg := bytes.NewBufferString(`<!DOCTYPE html>
			<html> <head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
			<title>采集点表文件上传</title>
			</head>  
			<body>`)
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		buf := bufio.NewReader(file)
		var cols *Columns
		iLine := 0
		total := 0
		bad := 0
		for line, err := buf.ReadString('\n'); len(line) > 0 || err == nil; line, err = buf.ReadString('\n') {
			iLine++
			if err != nil {
				break
			}
			if line[0] == '#' {
				line, err = buf.ReadString('\n')
				continue
			}
			line = strings.Trim(line, string([]byte{239, 187, 191})) //utf8 编码标识
			cols = ColumnsPosition(line)
			break
		}
		logs.Debug(cols.Address)
		conn := ctable.GetDB()
		if conn == nil {
			msg.WriteString("<h3>连接到数据库失败!</h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}
		
		//上传点表操作前,清空旧的采集点表
		//开启事务处理
		app := GetAppName()
		if app == "" {
			msg.WriteString("<h3>获取服务名错误! </h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}
		sqlString := "DELETE FROM point WHERE sn ='" + app + "' AND event = '0';"
		tx, err := conn.Begin()
		if err != nil {
			msg.WriteString("<h3>" + err.Error() + "</h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}

		_, err = tx.Exec(sqlString)
		if err != nil {
			logs.Info(sqlString)
			msg.WriteString("清空旧数据错误 ! " + err.Error() + "<br />")
			err = tx.Rollback()
			if err != nil {
				logs.Error("Rollback error.", err)
			}
		}

		for line, err := buf.ReadString('\n'); len(line) > 0 || err == nil; line, err = buf.ReadString('\n') {
			iLine++
			line = strings.TrimSpace(line)
			if len(line) == 0 || line[0] == '#' {
				continue
			}
			total++
			sqlString, err = FormatSqlFromLine(cols, line, app, false)
			if err != nil {
				msg.WriteString(fmt.Sprint("line:", iLine) + err.Error() + "<br />")
				bad++
				continue
			}
			_, err = tx.Exec(sqlString)
			if err != nil {
				logs.Info(sqlString)
				bad++
				msg.WriteString(fmt.Sprint("line:", iLine) + line + err.Error() + "<br />")
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			msg.WriteString("<h3>" + err.Error() + "</h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}
		msg.WriteString(fmt.Sprintf("<b> 上传总数:%d 上传成功:%d 上传失败:%d </b>", total, total-bad, bad))
		msg.WriteString("</body></html>")
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		msg.WriteTo(w)

		logs.Info("设置点表更新标志")
		pointsUpdate = true
		logs.Info("清空内存中点表")
		pointList.Reset() //清空内存中的点表
	} else {
		// 上传页面
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		html := `
<!DOCTYPE html>
<html> <head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<title>采集点表文件上传</title>
</head>  
<body>
格式:[ID],[Name],[Type],[Desc],[Address],[Base],[Coefficient]<br />
格式说明:<br />
ID: ID<br />
Name: 点名<br />
Type: 点类型<br />
Description:测点描述<br />
Address: 测点地址信息包括采集类型,采集地址等信息<br />
Base:基数<br />
Coefficient:系数<br />
ID,PointName,Address 至少要有两项<br /><br />
        <form action="/upload/points" method="post" enctype="multipart/form-data">
            <input type="file" name="file" value="" /> 
            <input type="submit" name="submit" />
		</form> 
</body></html>
`
		io.WriteString(w, html)
	}
}

//下载点表(导出数据库点表)
func HttpDownloadPoints(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	app := GetAppName()
	if app == "" {
		w.Write([]byte("不能获取有效的服务名!"))
		return
	}
	id, name, typ, desc, address, base, coef := -1, "", "AX", "", "", float64(0), float64(0)
	sqlString := "SELECT uid, pn, rt, ed, ad, fb, fk FROM point WHERE event='0' AND sn ='" + app + "';"
	db := ctable.GetDB()
	if db == nil {
		w.Write([]byte("不能连接到数据库!"))
		return
	}

	rows, err := db.Query(sqlString)
	if err != nil {
		logs.Error("Export point list,", err)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	w.Write([]byte("ID,Name,Type,Desc,Address,Base,Coefficient\n"))
	for rows.Next() {
		err = rows.Scan(&id, &name, &typ, &desc, &address, &base, &coef)
		if err != nil {
			logs.Info("SELECT FROM point:", err)
			continue
		}
		w.Write([]byte(fmt.Sprint(id, ",", name, ",", typ, ",", desc, ",", address, ",", base, ",", coef, "\n")))
	}
}

//上传控制点表
func HttpUploadEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		msg := bytes.NewBufferString(`<!DOCTYPE html>
			<html> <head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
			<title>采集点表文件上传</title>
			</head>  
			<body>`)
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		buf := bufio.NewReader(file)
		var cols *Columns
		iLine := 0
		total := 0
		bad := 0
		for line, err := buf.ReadString('\n'); len(line) > 0 || err == nil; line, err = buf.ReadString('\n') {
			iLine++
			if err != nil {
				break
			}
			if line[0] == '#' {
				line, err = buf.ReadString('\n')
				continue
			}
			line = strings.Trim(line, string([]byte{239, 187, 191})) //utf8 编码标识
			cols = ColumnsPosition(line)
			break
		}

		conn := ctable.GetDB()
		if conn == nil {
			msg.WriteString("<h3>连接到数据库失败!</h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}
		//上传点表操作前,清空旧的采集点表
		//开启事务处理
		app := GetAppName()
		if app == "" {
			msg.WriteString("<h3>获取服务名错误! </h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}
		sqlString := "DELETE FROM point WHERE sn ='" + app + "' AND event = '1';"
		tx, err := conn.Begin()
		if err != nil {
			msg.WriteString("<h3>" + err.Error() + "</h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}

		_, err = tx.Exec(sqlString)
		if err != nil {
			logs.Info(sqlString)
			msg.WriteString("清空旧数据错误 ! " + err.Error() + "<br />")
			err = tx.Rollback()
			if err != nil {
				logs.Error("Rollback error.", err)
			}
		}

		for line, err := buf.ReadString('\n'); len(line) > 0 || err == nil; line, err = buf.ReadString('\n') {
			iLine++
			line = strings.TrimSpace(line)
			if len(line) == 0 || line[0] == '#' {
				continue
			}
			total++
			sqlString, err = FormatSqlFromLine(cols, line, app, true)
			if err != nil {
				msg.WriteString(fmt.Sprint("line:", iLine) + err.Error() + "<br />")
				bad++
				continue
			}
			_, err = tx.Exec(sqlString)
			if err != nil {
				logs.Info(sqlString)
				bad++
				msg.WriteString(fmt.Sprint("line:", iLine) + line + err.Error() + "<br />")
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			msg.WriteString("<h3>" + err.Error() + "</h3>")
			msg.WriteString("</body></html>")
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			msg.WriteTo(w)
			return
		}
		msg.WriteString(fmt.Sprintf("<b> 上传总数:%d 上传成功:%d 上传失败:%d </b>", total, total-bad, bad))
		msg.WriteString("</body></html>")
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		msg.WriteTo(w)
		eventsUpdate = true
	} else {
		// 上传页面
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		html := `
<!DOCTYPE html>
<html> <head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<title>控制点表文件上传</title>
</head>  
<body>
格式:[ID],[Name],[Type],[Desc],[Address],[Base],[Coefficient]<br />
格式说明:<br />
ID: ID<br />
Name: 点名<br />
Type: 点类型<br />
Description:测点描述<br />
Address: 测点地址信息包括采集类型,采集地址等信息<br />
Base:基数<br />
Coefficient:系数<br />
ID,PointName,Address 至少要有两项<br /><br />
        <form action="/upload/events" method="post" enctype="multipart/form-data">
            <input type="file" name="file" value="" /> 
            <input type="submit" name="submit" />
		</form> 
</body></html>`
		io.WriteString(w, html)
	}

}

//下载控制点表
func HttpDownloadEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	app := GetAppName()
	if app == "" {
		w.Write([]byte("不能获取有效的服务名!"))
		return
	}
	id, name, typ, desc, address, base, coef := -1, "", "AX", "", "", float64(0), float64(0)
	sqlString := "SELECT uid, pn, rt, ed, ad, fb, fk FROM point WHERE event !='0' AND sn ='" + app + "';"
	db := ctable.GetDB()
	if db == nil {
		w.Write([]byte("不能连接到数据库!"))
		return
	}

	rows, err := db.Query(sqlString)
	if err != nil {
		logs.Error("Export point list,", err)
		w.Write([]byte(err.Error()))
		return
	}
	defer rows.Close()

	w.Write([]byte("ID,Name,Type,Desc,Address,Base,Coefficient\n"))
	for rows.Next() {
		err = rows.Scan(&id, &name, &typ, &desc, &address, &base, &coef)
		if err != nil {
			logs.Info("SELECT FROM point:", err)
			continue
		}
		w.Write([]byte(fmt.Sprint(id, ",", name, ",", typ, ",", desc, ",", address, ",", base, ",", coef, "\n")))
	}
}

//文件上传
func HttpUpload(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		if _, err := os.Stat("upload"); err != nil {
			os.Mkdir("upload", os.ModeDir)
		}
		file, h, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		f, err := os.Create("upload/" + h.Filename)
		defer f.Close()
		io.Copy(f, file)

		if s, ok := file.(Sizer); ok {
			str := fmt.Sprintf("上传成功, 共上传 %d Bytes.", s.Size())
			if s.Size() > 1024*1024*1024 {
				str = fmt.Sprintf("上传成功, 共上传 %.02f G Bytes.", float64(s.Size())/(1024*1024*1024))
			} else if s.Size() > 1024*1024 {
				str = fmt.Sprintf("上传成功, 共上传 %.02f M Bytes.", float64(s.Size())/(1024*1024))
			} else if s.Size() > 1024 {
				str = fmt.Sprintf("上传成功, 共上传 %.02f K Bytes.", float64(s.Size())/1024)
			}
			w.Write([]byte(str))
		} else {
			w.Write([]byte("上传成功"))
		}

		return
	} else {
		// 上传页面
		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
		html := `
<!DOCTYPE html>
<html> <head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<title>upload file</title>
</head>  
<body>
        <form action="/upload" method="post" enctype="multipart/form-data">
            <input type="file" name="file" value="" /> 
            <input type="submit" name="submit" />
		</form> 
</body></html>
`
		io.WriteString(w, html)
	}
}
func HttpCommand(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的

	JsonString := ""
	switch r.FormValue("command") {
	case "Stop":
		JsonString = "1"
		logs.Info("stop by web !")
		logs.GetBeeLogger().Flush()
		os.Exit(0)
	case "GetJsonPointsQuality":
		JsonString = GetJsonPointsQualityCount()
	case "GetJsonPoints":
		if len(r.FormValue("pointIds")) > 0 {
			strid := strings.Split(r.FormValue("pointIds"), ",")
			JsonString = GetJsonRtlPoints(strid)
		} else {
			JsonString = GetJsonPoints(0, 0, -1, "", "")
		}
	case "GetJsonPointsV2":
		start, limit := 0, 0
		var pointID int32 = -1
		var pointName string = ""
		var pointDesc = ""
		str := r.FormValue("start")
		if str != "" {
			st, err := strconv.Atoi(str)
			if err == nil {
				start = st
			}
		}
		str = r.FormValue("limit")
		if str != "" {
			li, err := strconv.Atoi(str)
			if err == nil {
				limit = li
			}
		}

		str = r.FormValue("pid")
		if str != "" {
			pid, err := strconv.Atoi(str)
			if err == nil {
				pointID = int32(pid)
			}
		}

		pointName = strings.ToUpper(r.FormValue("pname"))
		pointDesc = strings.ToUpper(r.FormValue("pdesc"))
		JsonString = GetJsonPointsV2(start, start+limit, pointID, pointName, pointDesc)
	case "Status":
		JsonString = "0"
	default:
		JsonString = "Unknown error."
	}
	w.Write([]byte(JsonString))
}

/***
 * 统计测点质量
 */
func GetJsonPointsQualityCount() string {
	quality := PointQualityCount{}
	points := GetPointList()
	for _, point := range points {
		if point.Status&OP_TIMEOUT != 0 {
			quality.Timeout++
		} else if point.Status&OP_BAD != 0 {
			quality.Bad++
		} else {
			quality.Good++
		}
	}
	quality.Total = len(points)
	if JsonString, err := json.Marshal(quality); err == nil {
		return string(JsonString)
	}
	return "{}"
}

func point2Row(point *Point) (row *Row) {
	row = new(Row)
	row.UID = point.UID
	row.AV = fmt.Sprintf("%.2f", point.Value)
	row.PN = point.Name
	row.ID = point.ID
	row.RT = point.Type
	if point.Status&OP_TIMEOUT != 0 {
		row.AS = 28
	} else if point.Status&OP_BAD != 0 {
		row.AS = 0
	} else {
		row.AS = 192
	}
	row.TM = point.Time
	row.PT = point.PointType
	row.FB = point.Base
	row.FK = point.Coefficient
	row.HW = point.Address
	row.ET = point.Extend
	row.RK = point.Remark
	row.ED = point.Desc
	row.CP = point.StationNumber
	row.UP = point.UpdateStatus
	return
}

func point2RowV2(point *Point) (row *Row) {
	row = new(Row)
	row.UID = point.UID
	row.AV = fmt.Sprintf("%.2f", point.Value)
	row.PN = point.Name
	row.ID = point.ID
	row.RT = point.Type
	row.AS = point.Status
	row.TM = point.Time
	row.PT = point.PointType
	row.FB = point.Base
	row.FK = point.Coefficient
	row.HW = point.Address
	row.ET = point.Extend
	row.RK = point.Remark
	row.ED = point.Desc
	row.CP = point.StationNumber
	row.UP = point.UpdateStatus
	return
}

func GetJsonRtlPoints(strid []string) string {
	webpoint := []RtlValue{}
	points := GetPointList()

	for _, point := range points {
		for _, szID := range strid {
			if szID == fmt.Sprint(point.ID) {
				rtlVal := RtlValue{}
				rtlVal.ID = int(point.ID)
				rtlVal.AV = fmt.Sprintf("%.3f", point.Value)
				if point.Status&OP_TIMEOUT != 0 {
					rtlVal.AS = "28"
				} else if point.Status&OP_BAD != 0 {
					rtlVal.AS = "0"
				} else {
					rtlVal.AS = "192"
				}
				rtlVal.TM = time.Unix(int64(point.Time), 0).Format("2006-01-02 15:04:05")
				webpoint = append(webpoint, rtlVal)
			}
		}
	}

	if JsonString, err := json.Marshal(webpoint); err == nil {
		return string(JsonString)
	}
	return "{}"
}

/***
 * 获取测点信息
 */
func GetJsonPoints(start int, limit int, pointID int32, pointName string, pointDesc string) string {
	json_points := Rows{}
	json_Count := 0
	json_Rows := []*Row{}
	points := GetPointList()

	if pointID < 0 && pointName == "" && pointDesc == "" {
		json_Count = len(points)
		if start == limit && limit == 0 {
			for _, point := range points {
				json_Rows = append(json_Rows, point2Row(point))
			}
		} else {
			if start < len(points) {
				if limit > len(points) {
					limit = len(points)
				}
				for _, point := range points[start:limit] {
					json_Rows = append(json_Rows, point2Row(point))
				}
			}
		}
	} else {
		idx := 0
		for _, point := range points {
			if pointID >= 0 {
				if point.ID == pointID {
					json_Rows = append(json_Rows, point2Row(point))
					idx++
					break
				}
				continue
			}

			if pointName != "" {
				if !strings.Contains(point.Name, pointName) {
					continue
				}
			}

			if pointDesc != "" {
				if !strings.Contains(strings.ToUpper(point.Desc), pointDesc) {
					continue
				}
			}

			if start == limit && limit == 0 {
				json_Rows = append(json_Rows, point2Row(point))

			} else if idx >= start && idx < limit {
				json_Rows = append(json_Rows, point2Row(point))
			}
			idx++
		}
		json_Count = idx
	}

	json_points.Count = json_Count
	json_points.Rows = json_Rows

	if json_Count > 0 {
		if JsonString, err := json.Marshal(json_points); err == nil {
			return string(JsonString)
		}
	}

	return `{"Count":1,"Rows":[{"UID":-1,"AS":768,"RT":-1,"PN":"没有数据","ED":"没有数据","ID":-1}]}`
}

/***
 * 获取测点信息
 */
func GetJsonPointsV2(start int, limit int, pointID int32, pointName string, pointDesc string) string {
	json_points := Rows{}
	json_Count := 0
	json_Rows := []*Row{}
	points := GetPointList()

	if pointID < 0 && pointName == "" && pointDesc == "" {
		json_Count = len(points)
		if start == limit && limit == 0 {
			for _, point := range points {
				json_Rows = append(json_Rows, point2RowV2(point))
			}
		} else {
			if start < len(points) {
				if limit > len(points) {
					limit = len(points)
				}
				for _, point := range points[start:limit] {
					json_Rows = append(json_Rows, point2RowV2(point))
				}
			}
		}
	} else {
		idx := 0
		for _, point := range points {
			if pointID >= 0 {
				if point.ID == pointID {
					json_Rows = append(json_Rows, point2RowV2(point))
					idx++
					break
				}
				continue
			}

			if pointName != "" {
				if !strings.Contains(point.Name, pointName) {
					continue
				}
			}

			if pointDesc != "" {
				if !strings.Contains(strings.ToUpper(point.Desc), pointDesc) {
					continue
				}
			}

			if start == limit && limit == 0 {
				json_Rows = append(json_Rows, point2RowV2(point))

			} else if idx >= start && idx < limit {
				json_Rows = append(json_Rows, point2RowV2(point))
			}
			idx++
		}
		json_Count = idx
	}

	json_points.Count = json_Count
	json_points.Rows = json_Rows

	if json_Count > 0 {
		if JsonString, err := json.Marshal(json_points); err == nil {
			return string(JsonString)
		}
	}

	return `{"Count":1,"Rows":[{"UID":-1,"AS":768,"RT":-1,"PN":"没有数据","ED":"没有数据","ID":-1}]}`
}
