//读取点表的配置文件
package points

import (
	"bufio"
	"opapi4/opevent/opeventer"
	"os"
	"scada/ctable"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

type Daser interface {
	Work(conn interface{}, control opeventer.Controler) (err error)
	Check()
}

const (
	AX_TYPE = 0
	DX_TYPE = 1
	I2_TYPE = 2
	I4_TYPE = 3
	R8_TYPE = 4
	//	AO_TYPE = 5 //输入点,开关量
	//	DO_TYPE = 6 //输入点,模拟量

	OP_GOOD    = 0x0    //< 点值为正常
	OP_BAD     = 0x0300 //< 点值的好坏，如果设置为1，则为坏值
	OP_TIMEOUT = 0x8000 //< 点值的超时，如果设置为1，则为超时

	OP_STATUS_DISCONNECT int32 = 0 //网络断开
	OP_STATUS_CONNECT    int32 = 1 //网络连上
	OP_STATUS_VAILD      int32 = 3 //数据有效

	STATUS_NO_UPDATED = 0
	STATUS_UPDATED    = 1

	STATUS_CONTROL_NULL = 0
	STATUS_CONTROL      = 1

	MAX_SEND_NUM = 20000
)

type Value struct {
	ControlStatus int
	Ack           *chan []byte
	PtrPoint      *Point
	Value         float64 //<值
	Status        uint16  //<状态
	Time          int64   //<时间
}

//测点配置项
type Point struct {
	UID           int         //<全局ID
	Name          string      //<点名
	ID            int32       //<ID
	Type          int32       //<OP类型
	Value         float64     //<值
	Status        uint16      //<状态
	Time          int64       //<时间
	PointType     int32       //<类型
	Base          float64     //<基数
	Coefficient   float64     //<系数
	StationNumber int32       //<站号
	Address       int32       //<地址
	Extend        string      //<扩展字段
	Remark        string      //<备注
	Desc          string      //<描述
	UpdateStatus  int32       //<更新状态	=0 时表示要发送数据
	VoidPtr       interface{} //存放扩展属性
}

var (
	pointList PointList //存放测点的数组
	//行解析成测点
	PFuncFormatPointFromLine func(idxID, idxAddr, idxType, pt_, idxPN, idxDesc, fk_, idxBase, idxExtend, idxRemark, idxExpr int, line string) *Point
	PFuncFormatAddress       func(point *Point, address string) error
	sendQueue                chan *Point //发送队列
	pointsUpdate             bool
	eventsUpdate             bool
	updateTime               int64 //仅在UDP模式下使用
)

func SetUpdateTime(t int64) {
	if t == 0 {
		t = time.Now().Unix()
	} else {
		updateTime = t
	}
}

func GetUpdateTime() int64 {
	return updateTime
}

//点表是否更新过
func IsUpdatePointList() bool {
	return pointsUpdate
}

//控制点表是否更新过
func IsUpdateEventList() bool {
	return eventsUpdate
}

//复位点表状态
func ResetPointListStatus() {
	pointsUpdate = false
}

//复位控制点表状态
func ResetEventListStatus() {
	eventsUpdate = false
}

func GetPointList() []*Point {
	return pointList.Get()
}

func Append(points ...*Point) {
	pointList.Append(points...)
}

type PointList struct {
	pointArray []*Point
	lock       *sync.RWMutex
}

func (this *PointList) Init() {
	this.lock = new(sync.RWMutex)
}

func (this *PointList) Reset() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.pointArray = make([]*Point, 0)
}

//追加
func (this *PointList) Append(points ...*Point) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.pointArray = append(this.pointArray, points...)
}

//获取slice
func (this *PointList) Get() []*Point {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.pointArray
}

func init() {
	pointList.Init()
	sendQueue = make(chan *Point, MAX_SEND_NUM)

	//PointIDMap = map[int]*Point{}
	PFuncFormatPointFromLine = FormatPoint
}

func SetTimeout() {
	points := GetPointList()
	for _, point := range points {
		point.Status |= OP_TIMEOUT
	}
}

func GoAutoPush() {
	for {
		points := GetPointList()
		t := time.Now().Unix() - 20

		for _, point := range points {
			if point.Time == 0 {
				continue
			}
			if point.Time > t {
				continue
			}
			if point.Status&OP_TIMEOUT == 0 {
				if point.Status&2 == 0 { //== 2 表示带时间的电量数据
					point.Time = t
				}
				sendQueue <- point
			}
		}
		time.Sleep(time.Second * 10)
	}
}

func SendPoint(point *Point) {
	sendQueue <- point
}

func GetQueuePoint() <-chan *Point {
	return sendQueue
}

func GetType(t string) int32 {
	s := strings.Trim(t, " ")
	if len(s) == 1 {
		tp, _ := strconv.Atoi(s)
		return int32(tp)
	} else {
		switch s {
		case "AX":
			return AX_TYPE
		case "DX":
			return DX_TYPE
		case "I2":
			return I2_TYPE
		case "I4":
			return I4_TYPE
		case "R8":
		}
	}
	return AX_TYPE
}

//把行格式化为测点
func FormatPoint(idxID, idxAddr, idxType, pt_, idxPN, idxDesc, fk_, idxBase, idxExtend, idxRemark, idxExpr int, line string) *Point {
	ss := strings.Split(line, ",")
	l := len(ss)
	if len(ss) < 2 {
		return nil
	}
	for i, _ := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
	point := new(Point)
	point.Status = OP_TIMEOUT
	point.UpdateStatus = 1
	point.Coefficient = 1
	point.Base = 0
	point.Address = 0
	if idxID >= 0 && idxID < l {
		id, _ := strconv.Atoi(strings.TrimSpace(ss[idxID]))
		point.ID = int32(id)
	}

	if idxPN >= 0 && idxPN < l {
		point.Name = strings.TrimSpace(ss[idxPN])
	}

	if idxDesc >= 0 && idxDesc < l {
		point.Desc = strings.TrimSpace(ss[idxDesc])
	}

	if idxExtend >= 0 && idxExtend < l {
		point.Extend = strings.TrimSpace(ss[idxExtend])
	}

	if idxRemark >= 0 && idxRemark < l {
		point.Remark = strings.TrimSpace(ss[idxRemark])
	}

	if fk_ >= 0 && fk_ < l {
		fk, err := strconv.ParseFloat(strings.TrimSpace(ss[fk_]), 64)
		if err == nil {
			point.Coefficient = fk
		}
	}

	if idxBase >= 0 && idxBase < l {
		fb, err := strconv.ParseFloat(strings.TrimSpace(ss[idxBase]), 64)
		if err == nil {
			point.Base = fb
		}
	}
	if idxAddr >= 0 && idxAddr < l {
		hw, _ := strconv.Atoi(strings.TrimSpace(ss[idxAddr]))
		point.Address = int32(hw)
	}

	if pt_ >= 0 && pt_ < l {
		pt, _ := strconv.Atoi(strings.TrimSpace(ss[pt_]))
		point.PointType = int32(pt)
	}

	if idxType >= 0 && idxType < l {
		point.Type = GetType(strings.TrimSpace(ss[idxType]))
	}

	return point
}

//从关系库中读取点表
func GetPointFromDB(servName string, isEvent bool) (arrays []*Point, err error) {
	strEvent := "='0' "
	id := -1
	name := ""
	typ := "AX"
	desc := ""
	address := ""
	base := float64(0)
	coef := float64(1)
	pt := ""

	if isEvent {
		strEvent = "!='0' "
	}
	sqlString := "SELECT uid, pn, rt, ed, ad, fb, fk, sr FROM point WHERE event" +
		strEvent + "AND sn ='" + servName + "';"
	db := ctable.GetDB()
	if db == nil {
		logs.Error("不能连接到数据库!")
		return
	}

	//logs.Debug("SQL:", sqlString)
	rows, err := db.Query(sqlString)
	if err != nil {
		logs.Error("Export point list,", err)
		return
	}
	defer rows.Close()

	arrays = make([]*Point, 0)
	for rows.Next() {
		err := rows.Scan(&id, &name, &typ, &desc, &address, &base, &coef, &pt)
		if err != nil {
			logs.Info("SELECT FROM point:", err)
			continue
		}
		point := new(Point)
		point.ID = int32(id)
		point.Name = name
		point.Status = OP_TIMEOUT
		point.Type = GetType(typ)
		point.Desc = desc

		pt = strings.TrimSpace(pt)
		if pt != "" {
			if strings.HasSuffix(address, ";") {
				address += pt
			} else {
				address += ";" + pt
			}
		}
		//logs.Debug("Address:", id, name, address, pt)

		if PFuncFormatAddress == nil {
			addr, err := strconv.Atoi(address)
			if err == nil {
				point.Address = int32(addr)
			}
		} else {
			err := PFuncFormatAddress(point, address)
			if err != nil {
				logs.Error(point.ID, point.Name, err)
				continue
			}
		}
		point.Base = base
		point.Coefficient = coef
		arrays = append(arrays, point)
	}
	return
}

//从文件中读取点组
func GetPointArray(path string) (points []*Point, err error) {
	if PFuncFormatPointFromLine == nil {
		PFuncFormatPointFromLine = FormatPoint
	}
	fe, err := os.Open(path)
	if err != nil {
		logs.Error("Faild to open file==>", path)
		return
	}
	defer fe.Close()

	idxID, idxAddr, idxType, pt_, idxPN, idxDesc, fk_, idxBase, idxExtend, idxRemark, idxExpr := -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
	buf := bufio.NewReader(fe)
	for line, err := buf.ReadString('\n'); len(line) > 0 || err == nil; line, err = buf.ReadString('\n') {
		if err != nil {
			break
		}
		if line[0] == '#' {
			line, err = buf.ReadString('\n')
			continue
		}
		line = strings.Trim(line, string([]byte{239, 187, 191})) //utf8 编码标识
		idxID, idxAddr, idxType, pt_, idxPN, idxDesc, fk_, idxBase, idxExtend, idxRemark, idxExpr = FormatColumnsPosition(line)
		break
	}

	points = make([]*Point, 0)
	for line, err := buf.ReadString('\n'); len(line) > 0 || err == nil; line, err = buf.ReadString('\n') {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		point := PFuncFormatPointFromLine(idxID, idxAddr, idxType, pt_, idxPN, idxDesc, fk_, idxBase, idxExtend, idxRemark, idxExpr, line)
		if point == nil {
			logs.Info("Parser error in line:", line+err.Error())
			continue
		}
		//logs.Info(point.Name, point.ID, point.Coefficient)
		points = append(points, point)
	}

	for i, point := range points {
		point.UID = i + 1
	}
	return
}

func (this *Point) Update(value float64, status uint16, time int64) {
	this.Value = value*this.Coefficient + this.Base
	this.Time = time
	if this.Type == DX_TYPE {
		this.Status = status | (uint16(value) & 1)
	} else {
		this.Status = status
	}
	this.UpdateStatus = 0
	sendQueue <- this
}

//读取每个字段的位置
func FormatColumnsPosition(columnsTitle string) (id, hw, rt, pt, pn, ed, fk, fb, et, rk, ex int) {
	line := strings.TrimSpace(columnsTitle)
	id, hw, rt, pt, pn, ed, fk, fb, et, rk, ex = -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
	ss := strings.Split(line, ",")
	for i := 0; i < len(ss); i++ {
		ss[i] = strings.TrimSpace(ss[i])
	}
	for i := 0; i < len(ss); i++ {
		switch ss[i] {
		case "ID":
			id = i
		case "PN":
			pn = i
		case "HW":
			hw = i
		case "PT":
			pt = i
		case "RT":
			rt = i
		case "ED":
			ed = i
		case "FK":
			fk = i
		case "FB":
			fb = i
		case "EX":
			ex = i
		case "ET":
			et = i
		case "RK":
			rk = i
		}
	}

	return
}
