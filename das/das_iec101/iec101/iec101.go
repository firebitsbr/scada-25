package iec101

/*
* IEC 101 为问答式
 */

import (
	"encoding/binary"
	"errors"
	"global"
	"io"
	"math"
	"opapi4/opevent/opeventer"
	"scada/points"
	"time"
	"wj/sock"

	"github.com/astaxie/beego/config"

	"github.com/astaxie/beego/logs"
)

const (
	ACD    = 0x20
	Level0 = 1
	Level1 = ACD
	Level2 = 2
)

var iecdas Iec101

type Iec101 struct {
	pointPath     string
	fcb           bool
	linkAddress   byte //链路地址
	commonAddress byte //公共地址
	pointSnMap    map[int32]*points.Point
}

//不可调用多次
func New(conf config.Configer) (das points.Daser, err error) {
	iecdas.pointPath = conf.DefaultString("service_name", points.GetAppName())
	err = iecdas.Initialize()
	if err != nil {
		return
	}
	iecdas.commonAddress = byte(conf.DefaultInt("iec101_common_address", 1))
	iecdas.linkAddress = byte(conf.DefaultInt("iec101_link_address", 1))
	das = &iecdas
	return
}

func (this *Iec101) Check() {
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
func (this *Iec101) LoadPointList() (err error) {
	pointArray, err := points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		logs.Error(err)
		return
	}
	this.pointSnMap = map[int32]*points.Point{}
	for _, point := range pointArray {
		this.pointSnMap[point.Address] = point
	}

	points.Append(pointArray...)
	return
}

func (this *Iec101) getFcb(c byte) byte {
	if this.fcb {
		c |= 0x20
	} else {
		c &= 0xDF
	}
	this.fcb = !this.fcb
	return c
}
func (this *Iec101) Initialize() (err error) {
	err = this.LoadPointList()
	go points.GoAutoPush()
	return
}

func CrcSum(data []byte) (cs byte) {
	cs = 0
	for _, b := range data {
		cs += b
	}
	return
}

//串口单线程跑,没有协程
func (this *Iec101) Work(conn interface{}, control opeventer.Controler) (err error) {
	this.Check()
	defer points.SetTimeout()
	if serial, ok := conn.(io.ReadWriteCloser); ok {
		err = this.TestFR(serial)
		if err != nil {
			err = errors.New("测试链路状态:" + err.Error())
			return
		}
		err = this.StartDT(serial)
		if err != nil {
			err = errors.New("复位链路状态:" + err.Error())
			return
		}
		t := time.Now()
		logs.Debug("发送总召唤命令")
		err = this.C_IC_NA_1(serial)
		if err != nil {
			return
		}
		for global.IsRunning() {
			this.Check()
			//总召唤
			data, err := this.Recv(serial)
			if err != nil {
				return err
			}
			level := this.doResponse(data)
			if time.Since(t)/time.Second > 600 {
				err = this.C_IC_NA_1(serial)
				if err != nil {
					return err
				}
			}

			select {
			case <-time.After(time.Millisecond * 500):
				err = this.requestClassData(serial, level)
				if err != nil {
					return err
				}
			case event := <-control.WaitEvent():
				control.DoMsg(serial, event, this)
			}
		}
	}
	return
}

//level = 0 表示数据要外部处理
//level = 1 表示有一级数据
//level = 2 表示有二级数据
func (this *Iec101) doResponse(data []byte) (level int) {
	level = Level2 //默认为2级数据
	if data[0] == 0x10 {
		if data[1]&ACD == ACD { //有一级数据
			//请求一级数据
			level = Level1
		}
	} else if data[0] == 0x68 {
		this.Route(data)
		if data[4]&ACD == ACD { //有一级数据
			//请求一级数据
			level = Level1
		}
		if data[6] == C_SC_NA { //遥控数据
			level |= Level0
			return
		}
	}
	return
}

func (this *Iec101) requestClassData(conn io.ReadWriter, level int) (err error) {
	if level&ACD != 0 {
		err = this.RequestClase1Data(conn)
	} else {
		err = this.RequestClase2Data(conn)
	}
	return
}

//请求2级数据
func (this *Iec101) RequestClase2Data(conn io.ReadWriter) (err error) {
	err = this.sendFixPacket(conn, 0x7B)
	if err != nil {
		return
	}
	return
}

//请求1级数据
func (this *Iec101) RequestClase1Data(conn io.ReadWriter) (err error) {
	err = this.sendFixPacket(conn, 0x7A)
	if err != nil {
		return
	}
	return
}

//发送一个完整数据包
func (this *Iec101) Send(conn io.ReadWriter, data []byte) (err error) {
	switch data[0] {
	case 0x10:
		data = append(data, CrcSum(data[1:3]), 0x16)
	case 0x68:
		l := len(data)
		data = append(data, CrcSum(data[4:l]), 0x16)
		data[1] = byte(l - 4)
		data[2] = data[1]
	}
	_, err = conn.Write(data)
	logs.Debug("S: % 02X", data)
	return
}

//接收一个完整
func (this *Iec101) Recv(conn io.ReadWriter) (data []byte, err error) {
	data = make([]byte, 1)
	_, err = sock.ReadTimeout(conn, data, 20*1000)
	if err != nil {
		return
	}
	switch data[0] {
	case 0x68:
		buf := make([]byte, 3)
		_, err = sock.ReadTimeout(conn, buf, 20*1000)
		if err != nil {
			return
		}
		data = append(data, buf...)
		if buf[0] != buf[1] {
			err = errors.New("Length error")
			return
		}
		buf = make([]byte, int(uint(buf[0]))+2)
		_, err = sock.ReadTimeout(conn, buf, 20*1000)
		if err != nil {
			return
		}
		data = append(data, buf...)
		l := len(buf)
		cs := CrcSum(buf[:l-2])
		if cs != buf[l-2] {
			err = errors.New("CRC ERROR.")
		}
	case 0x10:
		buf := make([]byte, 4)
		_, err = sock.ReadTimeout(conn, buf, 20*1000)
		if err != nil {
			return
		}
		data = append(data, buf...)
		if data[3] != (data[1] + data[2]) {
			err = errors.New("CRC ERROR.")
		}
	case 0xE5:
	default:
		err = errors.New("Packet error")
		return
	}
	logs.Debug("R: % 02X", data)
	return
}

//数据包路由选择
//只有首字符为0x68的才能有效
func (this *Iec101) Route(data []byte) (err error) {
	if len(data) == 0 {
		return
	}
	if data[0] != 0x68 {
		return
	}
	switch data[6] {
	case M_SP_NA: //单点信息
		this.M_SP_NA_1(data[6:])
	case M_ME_NA: //测量值, 规一化值
		this.M_ME_NA_1(data[6:])
	case M_ME_NC: //测量值, 短浮点数 0xd
		this.M_ME_NC_1(data[6:])
	case C_IC_NA: //总召唤
		switch data[8] {
		case 0x07:
			logs.Debug("收到总召唤激活确认帧")
		case 0x0A:
			logs.Debug("收到总召唤激活结束帧")
		default:
			logs.Debug("收到未知类型的总召唤应答帧")
		}
	}
	return
}

//编码信息体地址
func EncodeInfoAddress(address uint16) (ioa []byte) {
	ioa = make([]byte, 2)
	binary.LittleEndian.PutUint16(ioa, address)
	return
}

//解码信息体地址
func DecodeInfoAddress(data []byte) (a int, err error) {
	if len(data) < 2 {
		err = errors.New("length error")
		return
	}
	a = int(binary.LittleEndian.Uint16(data))
	return
}

func (this *Iec101) Fill(ctrl, tid, vsq, toc byte, ioa uint16, data []byte) (buf []byte) {
	buf = make([]byte, 10)
	buf[0] = 0x68
	buf[3] = 0x68
	buf[4] = this.getFcb(ctrl)
	buf[5] = this.linkAddress
	buf[6] = tid
	buf[7] = vsq
	buf[8] = toc
	buf[9] = this.commonAddress
	buf = append(buf, EncodeInfoAddress(ioa)...)
	buf = append(buf, data...)
	return
}

//单点遥控 (45)
func (this *Iec101) C_SC_NA_1(conn io.ReadWriter, point *points.Point) (err error) {
	ctrl := byte(0x73)
	tid := byte(C_SC_NA) //单点控制
	vsq := byte(1)       //VSQ
	toc := byte(0x6)     //传送原因
	ioa := uint16(point.Address)

	value := byte(point.Value) & 1
	value |= C_SC_NA_SELECT //预执行
	data := []byte{value}

	buf := this.Fill(ctrl, tid, vsq, toc, ioa, data)

	err = this.Send(conn, buf)
	logs.Notice("下发遥控预执行命令: ", point.Name, point.ID, point.Address, "value", value&1)

	for global.IsRunning() {
		select {
		case <-time.After(time.Second * 20):
			err = errors.New("接收遥控执行命令超时")
			logs.Notice(point.Name, point.ID, point.Address, err)
			return
		case <-time.After(time.Millisecond * 500):
			data, err = this.Recv(conn)
			if err != nil {
				return
			}
			level := this.doResponse(data)
			logs.Debug(value, level)
			if level&0x1 == Level0 {
				if data[0] != 0x68 {
					err = errors.New("Packet type error.")
					return
				}
				if len(data) < 15 {
					err = errors.New("Length error")
					return
				}

				if data[6] != C_SC_NA {
					err = errors.New("TID error")
					return
				}
				if value&0x80 != 0 { //预执行
					if data[8] == 0x7 {
						if data[12] == value {
							logs.Notice(point.Name, point.ID, point.Address, "收到遥控预执行命令返校确认")
							value &= 1
							data = []byte{value}
							buf = this.Fill(ctrl, tid, vsq, toc, ioa, data)
							logs.Notice("下发遥控执行命令: ", point.Name, point.ID, point.Address, "value", value&1)
							err = this.Send(conn, buf)
							if err != nil {
								return
							}
							continue
						} else {
							err = errors.New("收到遥控预执行命令返校确认不正确")
							logs.Notice(point.Name, point.ID, point.Address, err)
						}
						this.requestClassData(conn, level)
					} else if data[8] == 0xA {
						this.requestClassData(conn, level)
						if data[12] == value {
							err = errors.New("收到遥控预执行命令否定返校确认")
							logs.Notice(point.Name, point.ID, point.Address, err)
							return
						}
						err = errors.New("收到不确定的遥控预执行命令返校确认")
					}

				} else {
					this.requestClassData(conn, level)
					if data[8] == 0x7 {
						if data[12] == value {
							logs.Notice(point.Name, point.ID, point.Address, "收到遥控预执行命令返校确认")
							value &= 1
							data = []byte{value}
							buf = this.Fill(ctrl, tid, vsq, toc, ioa, data)
							logs.Notice("下发遥控执行命令: ", point.Name, point.ID, point.Address, "value", value&1)
							err = this.Send(conn, buf)
							if err != nil {
								return
							}
							continue
						} else {
							err = errors.New("收到遥控预执行命令返校确认不正确")
							logs.Notice(point.Name, point.ID, point.Address, err)
						}
					} else if data[8] == 0xA {
						if data[12] == value {
							logs.Notice(point.Name, point.ID, point.Address, "遥控成功")
							err = this.requestClassData(conn, level)
							return
						}
						err = errors.New("收到不确定的遥控执行命令返校确认")
					}

				}
			} else {
				err = this.requestClassData(conn, level)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

//发送总召唤
func (this *Iec101) C_IC_NA_1(conn io.ReadWriter) (err error) {
	ctrl := byte(0x73)
	tid := byte(C_IC_NA) //单点控制
	vsq := byte(1)       //VSQ
	toc := byte(0x6)     //传送原因
	ioa := uint16(0)
	data := []byte{0x14} //总召唤限定词（QOI）20（14H）
	buf := this.Fill(ctrl, tid, vsq, toc, ioa, data)

	err = this.Send(conn, buf)
	logs.Debug("发送总召唤命令")
	return
}

//只是单纯的ASDU部分
//单点信息 (1)
func (this *Iec101) M_SP_NA_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 4
	offset := 1
	t := time.Now().Unix()
	if isOrder { //地址是顺序累加的
		address, err := DecodeInfoAddress(asdu[idx:])
		if err != nil {
			return nil
		}
		idx += 2
		for i := 0; i < num && idx < l; i++ {
			//logs.Info(address+i, asdu[idx]&1)
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				point.Update(float64(asdu[idx]&1), 0, t)
			}
			idx += offset
		}
	} else {
		offset = 3
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			//logs.Info(address, asdu[idx]&1)
			if point, ok := this.pointSnMap[int32(address)]; ok {
				point.Update(float64(asdu[idx+2]&1), 0, t)
			}
			idx += offset
		}
	}
	return
}

//测量值, 规一化值 (9)
func (this *Iec101) M_ME_NA_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 4
	offset := 3
	t := time.Now().Unix()
	if isOrder {
		address, err := DecodeInfoAddress(asdu[idx:])
		if err != nil {
			return nil
		}
		idx += 2
		for i := 0; i < num && idx < l; i++ {
			logs.Info(address+i, binary.LittleEndian.Uint16(asdu[idx:]))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				point.Update(float64(binary.LittleEndian.Uint16(asdu[idx:])), 0, t)
			}
			idx += offset
		}
	} else {
		offset = 5
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			//logs.Info(address, binary.LittleEndian.Uint16(asdu[idx+2:]))
			if point, ok := this.pointSnMap[int32(address)]; ok {
				point.Update(float64(binary.LittleEndian.Uint16(asdu[idx+2:])), 0, t)
			}
			idx += offset
		}
	}
	return
}

//测量值,短浮点数 (13)
func (this *Iec101) M_ME_NC_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 4
	offset := 5
	t := time.Now().Unix()
	if isOrder {
		address, err := DecodeInfoAddress(asdu[idx:])
		if err != nil {
			return nil
		}
		idx += 2
		for i := 0; i < num && idx < l; i++ {
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				point.Update(float64(math.Float32frombits(binary.LittleEndian.Uint32(asdu[idx:]))), 0, t)
			}
			idx += offset
		}
	} else {
		offset = 7
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			if point, ok := this.pointSnMap[int32(address)]; ok {
				point.Update(float64(math.Float32frombits(binary.LittleEndian.Uint32(asdu[idx+2:]))), 0, t)
			}
			idx += offset
		}
	}
	return
}

//发送固定帧
func (this *Iec101) sendFixPacket(conn io.ReadWriter, lc byte) (err error) {
	buf := make([]byte, 3)
	buf[0] = 0x10
	if lc == 0x40 || lc == 0x49 {
		buf[1] = lc
	} else {
		buf[1] = this.getFcb(lc)
	}
	buf[2] = this.linkAddress
	err = this.Send(conn, buf)
	return
}

//接收固定帧
func (this *Iec101) recvFixPacket(conn io.ReadWriter) (ctrl byte, err error) {
	data, err := this.Recv(conn)
	if err != nil {
		return
	}
	if data[0] != 0x10 {
		err = errors.New("ERROR RESPONSE.")
		return
	}

	if len(data) != 5 {
		err = errors.New("ERROR RESPONSE LENGTH.")
		return
	}

	ctrl = data[1]
	return
}

//起始帧 (复位链路)
func (this *Iec101) StartDT(conn io.ReadWriter) (err error) {
	logs.Debug("链路状态复位")
	err = this.sendFixPacket(conn, 0x40)
	if err != nil {
		return
	}

	ctrl, err := this.recvFixPacket(conn)
	if ctrl&0x0f != 0 {
		err = errors.New("Control filed error. must be 0x20")
		return
	}

	logs.Debug("链路状态复位成功")
	return
}

//测试帧(链路状态)
func (this *Iec101) TestFR(conn io.ReadWriter) (err error) {
	logs.Debug("测试链路状态")
	err = this.sendFixPacket(conn, 0x49)
	if err != nil {
		return
	}

	ctrl, err := this.recvFixPacket(conn)
	if ctrl&0x0f != 0x0b {
		err = errors.New("Control filed error. must be 0x20")
		return
	}
	logs.Debug("链路状态测试成功")
	return
}
