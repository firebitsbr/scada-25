package iec104

import (
	"encoding/binary"
	"errors"
	"fmt"
	"global"
	"io"
	"log"
	"math"
	"opapi4/opevent/opeventer"
	"scada/points"
	"time"
	"wj/estring"
	"wj/exception"
	"wj/sock"

	"github.com/astaxie/beego/config"

	"github.com/astaxie/beego/logs"
)

var iecdas Iec104

type Iec104 struct {
	pointPath        string      //点表顺序
	sn               uint16      //发送序列号
	rn               uint16      //接收序列号
	commonAddress    uint16      //公共地址
	ackCmds          chan []byte //接收控制命令的缓存
	ackCmdMaxNum     int         //应答
	pointSnMap       map[int32]*points.Point
	enableC_CI_NA    bool  //是否开启电度数据读取
	c_ci_na_interval int64 //读取电度数据间隔 默认5分钟
}

//不可调用多次
func New(conf config.Configer) (das points.Daser, err error) {
	err = iecdas.Initialize()
	if err != nil {
		return
	}

	iecdas.pointPath = conf.DefaultString("service_name", points.GetAppName())
	iecdas.c_ci_na_interval = conf.DefaultInt64("iec104_c_ci_na_interval", 600)
	iecdas.enableC_CI_NA = conf.DefaultBool("iec104_c_ci_na_enable", false)
	iecdas.commonAddress = uint16(conf.DefaultInt("iec104_common_address", 1))
	if iecdas.c_ci_na_interval < 1 {
		iecdas.c_ci_na_interval = 2
	}
	das = &iecdas
	iecdas.LoadPointList()
	return
}

func (this *Iec104) Check() {
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
func (this *Iec104) LoadPointList() (err error) {
	pointArray, err := points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		logs.Error(err)
		return
	}
	this.pointSnMap = map[int32]*points.Point{}
	for _, point := range pointArray {
		this.pointSnMap[point.Address] = point
		//logs.Debug(point.Address, point.Name, point.ID)
	}

	points.Append(pointArray...)
	return
}

func (this *Iec104) Initialize() (err error) {
	this.ackCmdMaxNum = 100
	this.ackCmds = make(chan []byte, this.ackCmdMaxNum)
	err = this.LoadPointList()
	go points.GoAutoPush()
	return
}

//是否开启电度数据读取
func (this *Iec104) EnableC_CI_NA(b bool) {
	this.enableC_CI_NA = b
}
func (this *Iec104) pushCommand(cmd []byte) (err error) {
	if len(this.ackCmds) >= this.ackCmdMaxNum {
		err = errors.New("Queue is full")
		return
	}
	this.ackCmds <- cmd
	return
}

func (this *Iec104) Work(conn interface{}, control opeventer.Controler) (err error) {
	this.Check()
	defer points.SetTimeout()
	if tcpconn, ok := conn.(io.ReadWriteCloser); ok {
		//清空缓存
		for len(this.ackCmds) > 0 {
			<-this.ackCmds
		}

		over := make(chan bool, 1)
		go this.goRecvPacket(tcpconn, over)
		//发送起始帧
		err = this.StartDT(tcpconn)
		if err != nil {
			logs.Info(err)
			return
		}
		//总召
		err = this.C_IC_NA_1(tcpconn)
		if err != nil {
			logs.Info(err)
			return
		}

		if this.enableC_CI_NA {
			//请求一次电度数据
			err = this.C_CI_NA_1(tcpconn)
			if err != nil {
				logs.Info(err)
				return
			}
		}

		for global.IsRunning() {
			this.Check()
			select {
			case <-time.After(time.Second * time.Duration(this.c_ci_na_interval)): //每5分钟请求一次电度数据/* 60 * 5*/
				if this.enableC_CI_NA {
					//请求一次电度数据
					err = this.C_CI_NA_1(tcpconn)
					if err != nil {
						logs.Info(err)
						return
					}
				}
			case <-time.After(time.Second * 60 * 15): //15分钟发送一次总招
				err = this.C_IC_NA_1(tcpconn)
				if err != nil {
					logs.Info(err)
					return
				}
			case <-over:
				err = errors.New("over")
				return
			case data := <-this.ackCmds:
				fmt.Printf("忽略的指令: % 02X\n", data)
			case event := <-control.WaitEvent():
				control.DoMsg(tcpconn, event, this)
			}
		}
	}
	return
}

//发送一个完整数据包
func (this *Iec104) Send(conn io.ReadWriter, data []byte) (err error) {
	_, err = conn.Write(data)
	logs.Debug("S: % 02X", data)
	if data[2]&1 == 0 {
		this.sn++
	}
	return
}

//接收一个完整
func (this *Iec104) Recv(conn io.ReadWriter) (data []byte, err error) {
	data = make([]byte, 2)
	_, err = sock.ReadTimeout(conn, data, 60*1000)
	if err != nil {
		return
	}
	if data[0] != 0x68 {
		err = errors.New("IEC104 heard error")
		return
	}
	buf := make([]byte, int(uint(data[1])))
	_, err = sock.ReadTimeout(conn, buf, 60*1000)
	if err != nil {
		return
	}
	data = append(data, buf...)
	if data[2]&1 == 0 {
		this.rn++
	}
	logs.Debug("R: % 02X", data)
	return
}

//接收数据包协程
func (this *Iec104) goRecvPacket(conn io.ReadWriter, over chan bool) {
	defer func() {
		close(over) //相当于发送广播消息
		exception.CatchException()
	}()
	for global.IsRunning() {
		buf, err := this.Recv(conn)
		if err != nil {
			logs.Info(err)
			return
		}

		if len(buf) < 6 {
			logs.Info("package length error")
			return
		}

		if buf[0] != 0x68 {
			logs.Info("Packet head != 0x68")
			return
		}

		//数据包路由选择处理
		err = this.Route(conn, buf)
		if err != nil {
			logs.Info(err)
			return
		}
	}
}

//数据包路由选择
func (this *Iec104) Route(conn io.ReadWriter, data []byte) (err error) {
	switch data[2] & 3 { //控制域
	case 1: //S 帧 (数据监视帧)
		//未识别的数据全部放入到命令中
		err = this.pushCommand(data)
		if err != nil {
			return
		}
	case 3: //U 帧 (控制帧)
		switch data[2] {
		case TESTFR:
			err = this.sendFixPacket(conn, TESTFR_ACT)
		case STARTDT:
			err = this.sendFixPacket(conn, STARTDT_ACT)
		default:
			//未识别的数据全部放入到命令中
			err = this.pushCommand(data)
		}
		if err != nil {
			return
		}
	default: //信息帧
		if len(data) < 11 {
			return
		}
		switch data[6] { //类型标识
		case M_SP_NA: //单点信息
			err = this.M_SP_NA_1(data[6:])
		case M_ME_NA: //测量值, 规一化值
			err = this.M_ME_NA_1(data[6:])
		case M_ME_NC: //测量值, 短浮点数
			err = this.M_ME_NC_1(data[6:])
		case M_ME_ND: //不带品质描述的归一化值
			err = this.M_ME_ND_1(data[6:])
		case M_IT_NA: //累积量, 电度数据
			err = this.M_IT_NA_1(data[6:])
		case M_SP_TB: //带时标遥信
			err = this.M_SP_TB_1(data[6:])
		case C_IC_NA: //总召
			switch binary.LittleEndian.Uint16(data[8:]) {
			case 0x7:
				logs.Debug("收到总召激活确认帧")
			case 0xa:
				logs.Debug("收到总召结束帧")
			}
		case C_CI_NA: //电度召唤
			switch binary.LittleEndian.Uint16(data[8:]) {
			case 0x7:
				logs.Debug("收到召唤电度激活确认帧")
			case 0xa:
				logs.Debug("收到召唤电度结束帧")
			}
		case M_IT_TB: //带时标CP56Time2a的累计量
			err = this.M_IT_TB_1(data[6:])
		default:
			//未识别的数据全部放入到命令中
			err = this.pushCommand(data)
		}
		if err != nil {
			return
		}
		err = this.sendFixPacket(conn, S_ACK)
	}
	return
}

//编码信息体地址
func EncodeInfoAddress(address uint) (ioa []byte) {
	ioa = make([]byte, 3)
	ioa[0] = byte(address)
	ioa[1] = byte(uint(address) >> 8)
	ioa[2] = byte(uint(address) >> 16)
	return
}

//解码信息体地址
func DecodeInfoAddress(data []byte) (a int, err error) {
	if len(data) < 3 {
		err = errors.New("length error")
		return
	}
	a = int(uint(data[0]) | uint(data[1])<<8 | uint(data[2])<<16)
	return
}

//编码公共地址
func EncodeCommonAddress(address uint16) (coa []byte) {
	coa = make([]byte, 2)
	binary.LittleEndian.PutUint16(coa, address)
	return
}

func (this *Iec104) Fill(tid byte, vsq byte, toc uint16, ioa uint, data []byte) (buf []byte) {
	sn := this.sn << 1
	rn := this.rn << 1
	buf = make([]byte, 10)
	buf[0] = 0x68
	binary.LittleEndian.PutUint16(buf[2:], sn)
	binary.LittleEndian.PutUint16(buf[4:], rn)
	buf[6] = tid                                                  //类型标识
	buf[7] = vsq                                                  //VSQ
	binary.LittleEndian.PutUint16(buf[8:], toc)                   //传送原因
	buf = append(buf, EncodeCommonAddress(this.commonAddress)...) //公共地址
	buf = append(buf, EncodeInfoAddress(ioa)...)                  //信息体地址
	buf = append(buf, data...)
	buf[1] = byte(len(buf) - 2)
	return
}

//单点遥控 (45)
func (this *Iec104) C_SC_NA_1(conn io.ReadWriter, point *points.Point) (err error) {
	tid := byte(C_SC_NA) //单点控制
	vsq := byte(1)       //VSQ
	toc := uint16(0x6)   //传送原因
	ioa := uint(point.Address)

	value := byte(point.Value) & 1
	value |= C_SC_NA_SELECT //预执行
	data := []byte{value}

	buf := this.Fill(tid, vsq, toc, ioa, data)

	err = this.Send(conn, buf)
	logs.Notice("下发遥控预执行命令: ", point.Name, point.ID, point.Address, "value", value)

NEXT_FRAME:
	for {
		select {
		case <-time.After(time.Second * 10):
			err = errors.New("接收遥控预执行命令返校超时")
			logs.Notice(point.Name, point.ID, point.Address, err)
			return
		case cmd := <-this.ackCmds:
			if len(cmd) >= 16 {
				if cmd[6] == byte(C_SC_NA) {
					if cmd[8] == 0x07 {
						if cmd[15] == value {
							logs.Notice(point.Name, point.ID, point.Address, "收到遥控预执行命令返校确认")
							break NEXT_FRAME
						}
						err = errors.New("收到遥控预执行命令返校确认不正确")
						logs.Notice(point.Name, point.ID, point.Address, err)
					} else if cmd[8] == 0x0A {
						if cmd[15] == value {
							err = errors.New("收到遥控预执行命令否定返校确认")
							logs.Notice(point.Name, point.ID, point.Address, err)
							return
						}
						err = errors.New("收到不确定的遥控预执行命令返校确认")
					} else {
						err = errors.New("收到不确定的遥控预执行命令返校确认类型")
					}
					return
				}
			}
		}
	}

	value &= 1
	data = []byte{value}
	buf = this.Fill(tid, vsq, toc, ioa, data)

	err = this.Send(conn, buf)
	logs.Notice("下发遥控执行命令: ", point.Name, point.ID, point.Address, "value", value)

	for {
		select {
		case <-time.After(time.Second * 10):
			err = errors.New("接收遥控执行命令确认结束超时")
			logs.Notice(point.Name, point.ID, point.Address, err)
			return
		case cmd := <-this.ackCmds:
			log.Printf("C_SC_NA_1: % 02X\n", cmd)
			if len(cmd) >= 16 {
				if cmd[6] == byte(C_SC_NA) {
					if cmd[8] == 0x07 {
						if cmd[15] == value {
							logs.Notice(point.Name, point.ID, point.Address, "收到遥控执行命令确认")
							continue
						}
						err = errors.New("收到遥控执行命令确认不正确")
						logs.Notice(point.Name, point.ID, point.Address, err)
						continue
					} else if cmd[8] == 0x0A {
						if cmd[15] == value {
							logs.Notice(point.Name, point.ID, point.Address, "收到遥控执行命令确认结束")
							return
						}
						err = errors.New("收到遥控执行命令确认结束不正确")
						logs.Notice(point.Name, point.ID, point.Address, err)
					} else {
						err = errors.New(fmt.Sprint("收到遥控执行命令确认异常 CASE=", cmd[8]))
					}
					return
				}
			}
		}
	}
	return
}

//发送总召唤
func (this *Iec104) C_IC_NA_1(conn io.ReadWriter) (err error) {
	tid := byte(C_IC_NA) //单点控制
	vsq := byte(1)       //VSQ
	toc := uint16(0x6)   //传送原因
	ioa := uint(0)
	data := []byte{0x14} //总召唤限定词（QOI）20（14H）
	buf := this.Fill(tid, vsq, toc, ioa, data)

	err = this.Send(conn, buf)
	logs.Debug("发送总召唤命令")
	return
}

//发送电度数据召唤
func (this *Iec104) C_CI_NA_1(conn io.ReadWriter) (err error) {
	tid := byte(C_CI_NA) //单点控制
	vsq := byte(1)       //VSQ
	toc := uint16(0x6)   //传送原因
	ioa := uint(0)
	data := []byte{0x5} //总召唤限定词（QOI）20（14H）
	buf := this.Fill(tid, vsq, toc, ioa, data)

	err = this.Send(conn, buf)
	logs.Debug("发送总召唤命令")
	return
}

//只是单纯的ASDU部分
//单点信息 (1)
func (this *Iec104) M_SP_NA_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 1
	t := time.Now().Unix()
	logs.Debug("M_SP_NA:", num)
	if isOrder { //地址是顺序累加的
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			logs.Debug(address+i, asdu[idx])
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				logs.Debug(address+i, asdu[idx])
				if point.Type == points.DX_TYPE {
					point.Update(float64(asdu[idx]&1), 0, t)
				}
			}
			idx += offset
		}
	} else {
		offset = 4
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			logs.Debug(address, asdu[idx+3])
			if point, ok := this.pointSnMap[int32(address)]; ok {
				logs.Debug(address+i, asdu[idx])
				if point.Type == points.DX_TYPE {
					point.Update(float64(asdu[idx+3]&1), 0, t)
				}
			}
			idx += offset
		}
	}
	return
}

//测量值, 规一化值 (9)
func (this *Iec104) M_ME_NA_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 3
	t := time.Now().Unix()
	logs.Debug("M_ME_NA:", num)
	if isOrder {
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			logs.Debug(address+i, float64(int16(binary.LittleEndian.Uint16(asdu[idx:]))))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(int16(binary.LittleEndian.Uint16(asdu[idx:]))), 0, t)
				}
			}
			idx += offset
		}
	} else {
		offset = 6
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			logs.Debug(address, float64(int16(binary.LittleEndian.Uint16(asdu[idx+3:]))))
			if point, ok := this.pointSnMap[int32(address)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(int16(binary.LittleEndian.Uint16(asdu[idx+3:]))), 0, t)
				}
			}
			idx += offset
		}
	}
	return
}

//测量值,短浮点数 (13)
func (this *Iec104) M_ME_NC_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 5
	t := time.Now().Unix()
	logs.Debug("M_ME_NC:", num)
	if isOrder {
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			logs.Debug(address+i, float64(math.Float32frombits(binary.LittleEndian.Uint32(asdu[idx:]))))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(math.Float32frombits(binary.LittleEndian.Uint32(asdu[idx:]))), 0, t)
				}
			}
			idx += offset
		}
	} else {
		offset = 8
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			logs.Debug(address, float64(math.Float32frombits(binary.LittleEndian.Uint32(asdu[idx+3:]))))

			if point, ok := this.pointSnMap[int32(address)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(math.Float32frombits(binary.LittleEndian.Uint32(asdu[idx+3:]))), 0, t)
				}
			}
			idx += offset
		}
	}
	return
}

//不带品质描述的归一化值
func (this *Iec104) M_ME_ND_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 2
	t := time.Now().Unix()
	logs.Debug("M_ME_ND:", num)
	if isOrder {
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			logs.Debug(address+i, float64(binary.LittleEndian.Uint16(asdu[idx:])))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(binary.LittleEndian.Uint16(asdu[idx:])), 0, t)
				}
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
			//logs.Debug(address, float64(binary.LittleEndian.Uint16(asdu[idx+3:])))

			if point, ok := this.pointSnMap[int32(address)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(binary.LittleEndian.Uint16(asdu[idx+3:])), 0, t)
				}
			}
			idx += offset
		}
	}
	return
}

//带时标CP56Time2a的累计量
func (this *Iec104) M_IT_TB_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 12
	logs.Debug("M_IT_TB:", num)
	if isOrder {
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			t, _ := CP56Time2a(asdu[idx+5:])
			logs.Debug(address+i, float64(binary.LittleEndian.Uint32(asdu[idx:])), estring.TimeFormat(t.Unix()))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(binary.LittleEndian.Uint32(asdu[idx:])), 2, t.Unix())
				}
			}
			idx += offset
		}
	} else {
		offset = 15
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			t, _ := CP56Time2a(asdu[idx+8:])
			logs.Debug(address+i, float64(binary.LittleEndian.Uint32(asdu[idx+3:])), estring.TimeFormat(t.UTC().Unix()))

			if point, ok := this.pointSnMap[int32(address)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(binary.LittleEndian.Uint32(asdu[idx+3:])), 2, t.UTC().Unix())
				}
			}
			idx += offset
		}
	}
	return
}

//带时标的单点遥信
func (this *Iec104) M_SP_TB_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 8
	logs.Debug("M_SP_TB:", num)
	if isOrder { //地址是顺序累加的
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			t, _ := CP56Time2a(asdu[idx+1:])
			logs.Debug(address+i, asdu[idx+1], estring.TimeFormat(t.Unix()))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				if point.Type == points.DX_TYPE {
					point.Update(float64(asdu[idx]&1), 0, t.Unix())
				}
			}
			idx += offset
		}
	} else {
		offset = 11
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			t, _ := CP56Time2a(asdu[idx+4:])
			logs.Debug(address, asdu[idx+3], estring.TimeFormat(t.Unix()))
			if point, ok := this.pointSnMap[int32(address)]; ok {
				if point.Type == points.DX_TYPE {
					point.Update(float64(asdu[idx+3]&1), 0, t.Unix())
				}
			}
			idx += offset
		}
	}
	return
}

func (this *Iec104) M_IT_NA_1(asdu []byte) (err error) {
	isOrder := asdu[1]&0x80 == 0x80
	num := int(asdu[1] & 0x7f)
	l := len(asdu)
	idx := 6
	offset := 5
	t := time.Now().Unix()
	logs.Debug("M_IT_NA:", num)
	if isOrder {
		address, err := DecodeInfoAddress(asdu[6:])
		if err != nil {
			return nil
		}
		idx += 3
		for i := 0; i < num && idx < l; i++ {
			logs.Debug(address+i, binary.LittleEndian.Uint32(asdu[idx:]))
			if point, ok := this.pointSnMap[int32(address+i)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(binary.LittleEndian.Uint32(asdu[idx:])), 0, t)
				}
			}
			idx += offset
		}
	} else {
		offset = 8
		for i := 0; i < num && idx < l; i++ {
			address, err := DecodeInfoAddress(asdu[idx:])
			if err != nil {
				return nil
			}
			logs.Debug(address, binary.LittleEndian.Uint32(asdu[idx+3:]))
			if point, ok := this.pointSnMap[int32(address)]; ok {
				if point.Type != points.DX_TYPE {
					point.Update(float64(binary.LittleEndian.Uint32(asdu[idx+3:])), 0, t)
				}
			}
			idx += offset
		}
	}
	return
}

func CP56Time2a(buf []byte) (t time.Time, err error) {
	if len(buf) < 7 {
		err = errors.New("Time format error")
		return
	}
	timeFormat := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", uint16(buf[6])+2000, buf[5], buf[4]&0x1f, buf[3], buf[2], binary.LittleEndian.Uint16(buf)/1000)
	t, err = estring.Time(timeFormat)
	//fmt.Printf("time % 02x %s %s\n", buf[:8], timeFormat, t.UTC().String())
	return
}

//发送固定帧
func (this *Iec104) sendFixPacket(conn io.ReadWriter, lc byte) (err error) {
	if lc == STARTDT {
		this.rn = 0
		this.sn = 0
	}
	buf := make([]byte, 6)
	buf[0], buf[1] = 0x68, 0x04
	buf[2], buf[3] = lc, 0
	flag := lc & 0x3
	if flag == 1 {
		rn := this.rn << 1
		binary.LittleEndian.PutUint16(buf[4:], rn)
	} else if flag == 0x3 {
		buf[4], buf[5] = 0, 0
	}
	err = this.Send(conn, buf)
	return
}

//起始帧
func (this *Iec104) StartDT(conn io.ReadWriter) (err error) {
	logs.Debug("发送起始帧")
	err = this.sendFixPacket(conn, STARTDT)
	if err != nil {
		return
	}
	for {
		select {
		case <-time.After(time.Second * 10):
			err = errors.New("接收起始确认帧超时")
			return
		case cmd := <-this.ackCmds:
			if cmd[2] == STARTDT_ACT {
				logs.Debug("收到起始确认帧")
				return
			}
		}
	}
	return
}

//停止帧
func (this *Iec104) StopDT(conn io.ReadWriter) (err error) {
	logs.Debug("发送停止帧")
	err = this.sendFixPacket(conn, STOPDT)
	if err != nil {
		return
	}
	for {
		select {
		case <-time.After(time.Second * 10):
			err = errors.New("接收停止确认帧超时")
			return
		case cmd := <-this.ackCmds:
			if cmd[2] == STOPDT_ACT {
				logs.Debug("收到停止确认帧")
				return
			}
		}
	}
	return
}

//测试帧
func (this *Iec104) TestFR(conn io.ReadWriter) (err error) {
	logs.Debug("发送测试帧")
	err = this.sendFixPacket(conn, TESTFR)
	if err != nil {
		return
	}
	for {
		select {
		case <-time.After(time.Second * 10):
			err = errors.New("接收测试确认帧超时")
			return
		case cmd := <-this.ackCmds:
			if cmd[2] == TESTFR_ACT {
				logs.Debug("收到测试确认帧")
				return
			}
		}
	}
	return
}
