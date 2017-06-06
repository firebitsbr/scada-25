package ctrl_modbus_tcp

import (
	"encoding/json"
	"fmt"
	"modbus"
	"net/http"
	"scada/points"
	"strconv"
	"strings"
)

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

func (this *Control) HttpEvent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的

	JsonString := ""
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
	JsonString = this.GetJsonPoints(start, start+limit, pointID, pointName, pointDesc)
	w.Write([]byte(JsonString))
}

func (this *Control) GetJsonPoints(start int, limit int, pointID int32, pointName string, pointDesc string) string {
	json_points := Rows{}
	json_Count := 0
	json_Rows := []*Row{}
	points := this.pointArray
	pointID = -1
	pointName = ""
	pointDesc = ""

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
func point2Row(point *modbus.Point) (row *Row) {
	p, ok := point.Extend.(*points.Point)
	row = new(Row)
	row.AV = fmt.Sprintf("%.2f", point.Value)
	if ok {
		row.PN = p.Name
		row.ID = p.ID
		row.AS = p.Status
		row.RT = p.Type
		row.ET = p.Extend
		row.RK = p.Remark
		row.ED = p.Desc
		row.CP = p.StationNumber
		row.UP = p.UpdateStatus
	}
	row.TM = point.Time
	return
}
