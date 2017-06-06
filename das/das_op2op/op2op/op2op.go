package op2op

import (
	"global"
	"opapi3"

	"opapi4/opevent/opeventer"
	"scada/points"
	"time"

	"github.com/astaxie/beego/logs"
)

type Op2op struct {
	isSyncID   bool
	pointPath  string
	pointArray []*points.Point
	conn       *opapi3.OPConnect
}

func New(pointList string) (das points.Daser, err error) {
	d := new(Op2op)

	d.pointPath = pointList
	err = d.LoadPointList()
	das = d
	return
}

func (this *Op2op) Check() {
	if points.IsUpdatePointList() {
		logs.Notice("正在重新加载点表")
		err := this.LoadPointList()
		if err != nil {
			logs.Error("动态加载点表错误", err)
		} else {
			logs.Notice("正在重新加载点表完成.")
		}
		points.ResetPointListStatus()
	}
}

//ID 同步
func (this *Op2op) SyncID(conn *opapi3.OPConnect, names []string, points []*points.Point) (err error) {
	ids, err := conn.GetIdArrayByNameArray(names)
	if err != nil {
		return
	}
	for i, id := range ids {
		points[i].Address = int32(id)
	}
	return
}

func (this *Op2op) Sync() (err error) {
	//1次同步1000点
	idx := 0
	base := 0
	const Max = 1000
	logs.Info("开始同步数据库点表")
	names := make([]string, len(this.pointArray))

	for i, point := range this.pointArray {
		names[i] = point.Name
	}
	idx = len(this.pointArray)
	for i := 0; i < idx; i += Max {
		base = i + Max
		if base > idx {
			base = idx
		}
		for {
			err := this.SyncID(this.conn, names[i:base], this.pointArray[i:base])
			if err == nil {
				break
			}
			logs.Warn("同步点表失败:", err)
			time.Sleep(time.Second * 5)
		}
	}
	this.isSyncID = true
	logs.Info("数据库同步点表完成")
	return
}

//加载点表
func (this *Op2op) LoadPointList() (err error) {
	this.isSyncID = false
	this.pointArray, err = points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		logs.Error(err)
		return
	}

	points.Append(this.pointArray...)
	if this.conn != nil {
		this.Sync()
	}
	return err
}

//工作主函数
func (this *Op2op) Work(conn interface{}, control opeventer.Controler) (err error) {
	if opconn, ok := conn.(*opapi3.OPConnect); ok {
		this.conn = opconn
		for global.IsRunning() {
			this.Check()
			if !this.isSyncID {
				this.Sync()
			}

			t, er := this.conn.GetSystemTime()
			if er != 0 {
				time.Sleep(time.Second * 10)
				continue
			}

			points.SetUpdateTime(int64(t))

			pos := 0
			for i := 0; i < len(this.pointArray); i += 1000 {
				pos = i + 1000
				if pos > len(this.pointArray) {
					pos = len(this.pointArray)
				}
				this.Update(this.pointArray[i:pos])
			}

			time.Sleep(time.Second)
		}
	}
	return
}

func (this *Op2op) Update(points []*points.Point) {
	l := len(points)
	if l == 0 {
		return
	}
	ids := make([]int32, l)
	for i, point := range points {
		ids[i] = point.Address
	}
	vals, err := this.conn.GetRTValueArrayByIDArray(ids)
	if err != nil {
		logs.Debug("Get realtime value:", err)
		return
	}
	for i, point := range points {
		point.Update(vals[i].Value, vals[i].Status, vals[i].Time)
	}
}
