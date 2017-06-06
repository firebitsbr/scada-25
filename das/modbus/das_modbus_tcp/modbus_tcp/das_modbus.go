package modbustcp

import (
	"io"
	"modbus"
	"modbus/tcp_master"
	"opapi4/opevent/opeventer"
	"scada/points"
	"strconv"
	"strings"

	"github.com/astaxie/beego/config"

	"github.com/astaxie/beego/logs"
)

type Modbus struct {
	ModbusTcp *modbustcpmaster.TcpMaster
	pointPath string
}

func New(conf config.Configer) (das points.Daser, err error) {
	d := new(Modbus)

	//points.PFuncFormatPointFromLine = FormatPoint
	//pointArray, err := points.GetPointArray(pointList)
	d.pointPath = conf.DefaultString("service_name", points.GetAppName())
	err = d.LoadPointList()
	das = d
	return
}

func (this *Modbus) Check() {
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

func (this *Modbus) LoadPointList() (err error) {
	points.PFuncFormatAddress = FormatAddress
	pointArray, err := points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		logs.Error(err)
		return
	}

	this.ModbusTcp = new(modbustcpmaster.TcpMaster)
	this.ModbusTcp.Initialize()

	points.Append(pointArray...)
	for _, point := range pointArray {
		p := new(modbus.Point)
		p.DeviceID = byte(point.StationNumber / 100)
		p.FuncCode = byte(point.StationNumber % 100)
		p.Address = uint16(point.Address)
		p.Type = byte(point.PointType)
		p.Extend = point
		point.VoidPtr = p
		this.ModbusTcp.AddPoint(p)
	}
	this.ModbusTcp.InitGroup()
	//logs.Info(len(this.ModbusTcp.PointArray), len(this.ModbusTcp.ModbusGroup))
	return err
}

func (this *Modbus) Work(conn interface{}, control opeventer.Controler) (err error) {
	this.Check()
	if tcpconn, ok := conn.(io.ReadWriter); ok {
		var data []byte
		for _, pModbus := range this.ModbusTcp.ModbusGroup {
			if pModbus.GetCount() == 0 {
			}
			cmd := pModbus.MakeRequest()
			_, err = this.ModbusTcp.Write(tcpconn, cmd)
			if err != nil {
				break
			}
			data, err = this.ModbusTcp.Read(tcpconn)
			if err != nil {
				logs.Info(err)
				if strings.Index(strings.ToLower(err.Error()), "timeout") >= 0 {
					err = nil //超时, 直接查询下一组操作
					continue
				} else {
					break
				}
			}
			err = pModbus.DoResponse(data)
			if err != nil {
				return
			}

			if control != nil {
				if control.IsHaveEvent() {
					control.DoEvent(tcpconn, nil)
				}
			}
		}
		this.Update()
	}
	return
}

func (this *Modbus) Update() {
	for _, point := range this.ModbusTcp.PointArray {
		if point.IsUpdate {
			if p, ok := point.Extend.(*points.Point); ok {
				//logs.Debug(p.Name, point.Value, point.Time)
				p.Update(point.Value, 0, point.Time)
			}
		}
	}
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
			point.Remark = "COILS"
		} else {
			point.PointType = modbus.MODBUS_INT16
			point.Remark = "INT16"
		}
	case 2:
		id, code, addr, err := modbus.GetModbusInfo(ss[0])
		if err != nil {
			return err
		}
		point.StationNumber = int32(uint(id))*100 + int32(code)
		point.Address = int32(addr)
		point.PointType = int32(modbus.GetModbusType(ss[1]))
		point.Remark = ss[1]
	}
	return
}

/*	格式化行的数据(即把每行解析成一个测点配置)
*
 */
func FormatPoint(id_, hw_, rt_, pt_, pn_, ed_, fk_, fb_, et_, rk_, expr_ int, line string) *points.Point {
	line = strings.Replace(line, "\r", "", -1)
	line = strings.Replace(line, "\n", "", -1)
	line = strings.TrimSpace(line)
	ss := strings.Split(line, ",")
	leng := len(ss)
	if len(ss) < 2 {
		return nil
	}
	for i, _ := range ss {
		ss[i] = strings.TrimSpace(ss[i])
	}
	point := new(points.Point)
	point.Status = points.OP_TIMEOUT
	point.UpdateStatus = 1
	point.Coefficient = 1
	point.Base = 0
	point.Address = 0
	//logs.Info("ID:", id_)
	if id_ >= 0 && id_ < leng {
		id, _ := strconv.Atoi(strings.TrimSpace(ss[id_]))
		point.ID = int32(id)
	}

	if pn_ >= 0 && pn_ < leng {
		point.Name = strings.TrimSpace(ss[pn_])
	}

	if ed_ >= 0 && ed_ < leng {
		point.Desc = strings.TrimSpace(ss[ed_])
	}

	if et_ >= 0 && et_ < leng {
		point.Extend = strings.TrimSpace(ss[et_])
	}

	if rk_ >= 0 && rk_ < leng {
		point.Remark = strings.TrimSpace(ss[rk_])
	}

	if fk_ >= 0 && fk_ < leng {
		fk, err := strconv.ParseFloat(strings.TrimSpace(ss[fk_]), 64)
		if err == nil {
			point.Coefficient = fk
		}
	}

	if fb_ >= 0 && fb_ < leng {
		fb, err := strconv.ParseFloat(strings.TrimSpace(ss[fb_]), 64)
		if err == nil {
			point.Base = fb
		}
	}
	if hw_ >= 0 && hw_ < leng {
		var err error
		id, code, address, err := modbus.GetModbusInfo(ss[hw_])
		if err != nil {
			return nil
		}
		point.StationNumber = int32(uint(id))*100 + int32(code)
		point.Address = int32(address)
	} else {
		logs.Warn(line, " Failed to parser. hw is error")
		return nil
	}

	if pt_ >= 0 && pt_ < leng {
		point.PointType = int32(modbus.GetModbusType(ss[pt_]))
	} else {
		logs.Warn(line, " Failed to parser. Point's type is error")
		return nil
	}

	if rt_ >= 0 && rt_ < leng {
		point.Type = points.GetType(ss[rt_])
	}

	point.Remark = ss[hw_] + "; " + ss[pt_]
	return point
}
