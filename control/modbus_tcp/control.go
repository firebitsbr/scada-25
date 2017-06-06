package ctrl_modbus_tcp

import (
	"errors"
	"io"
	"modbus"
	"opapi4"
	"opapi4/opevent"
	"opapi4/opevent/opeventer"
	"scada/points"
	"strings"
	"time"
	"web"
	"wj/errs"
	"wj/sock"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

type Control struct {
	opevent.OPEvent
	PointIDMap map[int]*modbus.Point
	pointArray []*modbus.Point
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
	web.HandleFunc("/event", c.HttpEvent)

	err = c.Initialize(points.GetAppName() /*conf.String("control.points")*/)
	return
}

func FormatAddress(point *points.Point, address string) (err error) {
	ss := strings.Split(address, ";")
	switch len(ss) {
	case 1:
		id, code, addr, err := modbus.GetModbusInfo(ss[0])
		if err != nil {
			return err
		}
		point.StationNumber = int32(uint(id))*100 + int32(code)
		point.Address = int32(addr)
		if point.Type == points.DX_TYPE {
			point.PointType = modbus.MODBUS_COILS
		} else {
			point.PointType = modbus.MODBUS_INT16
		}
	case 2:
		id, code, addr, err := modbus.GetModbusInfo(ss[0])
		if err != nil {
			return err
		}
		point.StationNumber = int32(uint(id))*100 + int32(code)
		point.Address = int32(addr)
		point.PointType = int32(modbus.GetModbusType(ss[1]))
	}
	return
}

func (this *Control) Initialize(pointList string) (err error) {
	err = this.OPEvent.Initialize(200)
	if err != nil {
		return
	}

	this.PointIDMap = map[int]*modbus.Point{}

	//加载控制点表
	//	points.PFuncFormatPointFromLine = modbustcp.FormatPoint
	//	pointArray, err := points.GetPointArray(pointList)
	points.PFuncFormatAddress = FormatAddress
	pointArray, err := points.GetPointFromDB(pointList, true)
	if err != nil {
		return
	}

	this.pointArray = nil

	//ID映射控制点表
	for _, point := range pointArray {
		p := new(modbus.Point)
		p.DeviceID = byte(point.StationNumber / 100)
		p.FuncCode = byte(point.StationNumber % 100)
		p.Address = uint16(point.Address)
		p.Type = byte(point.PointType)
		p.Extend = point.Name
		this.PointIDMap[int(point.ID)] = p
		this.pointArray = append(this.pointArray, p)
	}

	//生成订阅清单
	ids := make([]int, 0, len(this.PointIDMap))
	for id := range this.PointIDMap {
		ids = append(ids, id)
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

func (this *Control) DoEvent(w io.ReadWriter, extend interface{}) (err error) {
	for len(this.InformQueue) > 0 {
		msg := <-this.WaitEvent()
		err = this.DoMsg(w, msg, extend)
		if err != nil {
			return
		}
	}
	return
}

func (this *Control) DoMsg(w io.ReadWriter, e *opeventer.Event, extend interface{}) (err error) {
	logs.Info("Control Message:", e.ID, e.Value, e.Status, e.Time, time.Unix(int64(e.Time), 0).Format("2006-01-02 15:04:05"))
	if point, ok := this.PointIDMap[int(e.ID)]; ok {
		point.Value = e.Value
		data, er := modbus.DowithModbusControlCmd(point)
		if er != nil {
			logs.Info("不能生成控制报文 ID:", e.ID, e.Value, point.DeviceID, er)
			return
		}
		logs.Info("生成控制报文:%d % 02X \n", e.ID, data)
		err = this.SetValue(w, data)

		if err != nil {
			this.CtrlFeedback(int(e.ID), int(e.Time), opapi4.CONTROL_BAD, e.Value)
			//控制失败
			if strings.Contains(err.Error(), "net error") {
				return
			}
			err = nil
		} else {
			this.CtrlFeedback(int(e.ID), int(e.Time), opapi4.CONTROL_GOOD, e.Value)
		}

	} else {
		logs.Warn("控制点不存在", e.ID, e.Value)
	}
	return
}

func (this *Control) SetValue(conn io.ReadWriter, data []byte) (err error) {
	head := make([]byte, 6)
	head[0], head[1] = 'c', 'l'
	head[2], head[3] = 0, 0
	head[4] = byte(len(data) >> 8)
	head[5] = byte(len(data))
	sbuf := make([]byte, 6)
	copy(sbuf, head)
	sbuf = append(sbuf, data...)

	_, err = conn.Write(sbuf)
	if err != nil {
		err = errors.New("net error" + err.Error())
		return
	}

	buf := make([]byte, 256)
	_, err = sock.ReadTimeout(conn, buf[:6], modbus.TCP_SLAVE_DEFAULT_TIMEOUT)
	if err != nil {
		err = errors.New("net error" + err.Error())
		return
	}
	l := uint16(buf[4]<<8) | uint16(buf[5])
	if l > uint16(len(buf)) {
		return errors.New("net error.")
	}
	_, err = sock.ReadTimeout(conn, buf[:l], modbus.TCP_SLAVE_DEFAULT_TIMEOUT)
	if l < 6 {
		err = errors.New("net error. Response error")
		return
	}

	if data[0] == buf[0] && data[1] == buf[1] && data[2] == buf[2] && data[3] == buf[3] {
		switch data[1] {
		case 5, 6:
			if !(data[4] == buf[4] && data[5] == buf[5]) {
				err = errors.New("Control error")
			}
		case 15, 16:
			if !(buf[4] > 0 || buf[5] > 0) {
				err = errors.New("Control error")
			}
		default:
			err = errors.New("Function code error")
		}
	}
	return
}
