package ctrl_iec101

import (
	"errors"
	"io"
	"opapi4"
	"opapi4/opevent"
	"opapi4/opevent/opeventer"
	"scada/control/iec101/interface"
	"scada/points"
	"strings"
	"time"
	"wj/errs"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

type Control struct {
	opevent.OPEvent
	PointIDMap map[int32]*points.Point
	pointArray []*points.Point
}

func New(conf config.Configer) (c *Control, err error) {
	c = new(Control)

	if !conf.DefaultBool("control_enable", false) {
		c.OPEvent.Initialize(0)
		return
	}

	c.Address = conf.String("control_address")
	c.User = conf.String("control_user_name")
	c.Pwd = conf.String("control_user_password")
	c.Option = 0
	c.Timeout = 60
	if c.Address == "" {
		err = errs.New("The address of control server is empty.")
		return
	}

	err = c.Initialize(points.GetAppName() /*conf.String("control_points")*/)
	return
}

func (this *Control) Initialize(pointList string) (err error) {
	err = this.OPEvent.Initialize(200)
	if err != nil {
		return
	}

	this.PointIDMap = map[int32]*points.Point{}

	//加载控制点表
	this.pointArray, err = points.GetPointFromDB(pointList, true)
	if err != nil {
		err = errs.New(err.Error())
		return
	}

	//ID映射控制点表
	for _, point := range this.pointArray {
		if point.Type != points.DX_TYPE {
			logs.Error("控制只支持开关量", point.Name, point.ID)
			continue
		}
		this.PointIDMap[point.ID] = point
	}

	//生成订阅清单
	ids := make([]int, 0, len(this.PointIDMap))
	for id := range this.PointIDMap {
		ids = append(ids, int(id))
	}

	if len(ids) == 0 {
		err = errs.New("没有订阅清单")
		return
	}

	logs.Info("控制订阅清单数量", len(ids))

	//检查连接
	logs.Notice("检查订阅连接是否正常")
	for this.GetStatus() != 0 {
		logs.Error("Retry connect to ", this.Address, this.User)
		time.Sleep(time.Second * 5)
	}
	logs.Notice("订阅连接正常")
	//订阅控制信息
	logs.Notice("正在订阅控制数据")
	for {
		event := opapi4.NewCtrlEventByIDs(this.OPConnectSession, ids, 200)
		if event == nil {
			logs.Error("订阅控制数据失败,请检查点表是否正确")
			time.Sleep(time.Second * 5)
			continue
		}
		this.SetEvent(event)
		break
	}
	logs.Notice("控制数据订阅完成.")
	return
}

//处理事件
func (this *Control) DoEvent(conn io.ReadWriter, extend interface{}) (err error) {
	for len(this.InformQueue) > 0 {
		msg := <-this.WaitEvent()
		err = this.DoMsg(conn, msg, extend)
		if err != nil {
			return
		}
	}
	return
}

//处理控制消息
func (this *Control) DoMsg(conn io.ReadWriter, e *opeventer.Event, extend interface{}) (err error) {
	logs.Info("Control Message:", e.ID, e.Value, e.Status, e.Time, time.Unix(int64(e.Time), 0).Format("2006-01-02 15:04:05"))
	if point, ok := this.PointIDMap[e.ID]; ok {
		point.Value = e.Value
		err = this.SetValue(conn, point, extend)

		if err != nil {
			this.CtrlFeedback(int(e.ID), int(e.Time), opapi4.CONTROL_BAD, e.Value)
			logs.Notice(point.Name, point.ID, "控制失败", err)
			//控制失败
			if strings.Contains(err.Error(), "net error") {
				return
			}
			err = nil
		} else {
			this.CtrlFeedback(int(e.ID), int(e.Time), opapi4.CONTROL_GOOD, e.Value)
			logs.Notice(point.Name, point.ID, "控制成功")
		}

	} else {
		logs.Warn("控制点不存在", e.ID, e.Value)
	}
	return
}

//修改值
func (this *Control) SetValue(conn io.ReadWriter, point *points.Point, extend interface{}) (err error) {
	if iec, ok := extend.(iec101.IECer); ok {
		return iec.C_SC_NA_1(conn, point)
	}
	err = errors.New("Type error.")
	return
}
