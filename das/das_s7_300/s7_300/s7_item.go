package s7_300

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/astaxie/beego/logs"
)

type Item struct {
	MemoryType byte
	DataType   byte
	DBBlock    uint16 //only for DB, DB1.B0.64
	Address    uint16 //开始地址
	Length     uint16 //计算请求个数
	PointArray []*S7Point
}

//添加一个测点到ITEM中
//err = nil, 表示成功, 否则表示不满足条件
func (this *Item) AddPoint(point *S7Point) (err error) {
	//log.Println(point.Address, point.MemoryType, point.DataType)
	if point == nil {
		return nil
	}
	if len(this.PointArray) == 0 {
		this.MemoryType = point.MemoryType
		this.DBBlock = point.DBBlock
		this.Address = point.Address
		this.DataType = point.DataType
		switch point.DataType {
		case DataType_B, DataType_C, DataType_X:
			this.Length += 1
		case DataType_D, DataType_DI, DataType_REAL:
			this.Length += 4
		case DataType_W, DataType_I:
			this.Length += 2
		default:
			this.Length += 1
		}
	} else {
		if this.MemoryType != point.MemoryType {
			return errors.New("Error memory type")
		}
		if this.DBBlock != point.DBBlock {
			return errors.New("Error block")
		}

		//判断地址是否在范围之内
		if point.Address-this.Address >= this.Length {
			return errors.New("Error address")
		}
	}
	this.PointArray = append(this.PointArray, point)
	//logs.Debug(this.PointArray, point.Address, this.Address, this.Length)
	return
}

func (this *Item) MakeRequest() (param []byte) {
	param = make([]byte, 12)
	param[0] = 0x12 //Variable specification //表示读取的是变量???
	param[1] = 0x0a //Length of following address specification //标识这个ITEM的长度???
	param[2] = 0x10 //Syntax ID S7ANY(0x10) //固定的???
	param[3] = 0x02 //Transport size BYTE(2) //标识传输的单位 2: 表示为字节

	//标识要读取的字节长度
	param[4] = byte(this.Length >> 8)  //读取的长度(高字节) //Big-Endian
	param[5] = byte(this.Length)       //读取的长度(低字节)
	param[6] = byte(this.DBBlock >> 8) // Only for db?
	param[7] = byte(this.DBBlock)      //

	addr := uint(this.Address * 8)
	switch this.MemoryType {
	case MemoryType_C:
		param[8] = 0x1c //地址不需要 * 8
		param[9] = byte(this.Address >> 16)
		param[10] = byte(this.Address >> 8)
		param[11] = byte(this.Address)
	case MemoryType_DB:
		param[8] = 0x84
		param[9] = byte(addr >> 16)
		param[10] = byte(addr >> 8)
		param[11] = byte(addr)
	case MemoryType_E, MemoryType_I: // 截取报文发现这两个是一个类型
		param[8] = 0x81
		param[9] = byte(addr >> 16)
		param[10] = byte(addr >> 8)
		param[11] = byte(addr)
	case MemoryType_M, MemoryType_F: // 截取报文发现这两个是一个类型
		param[8] = 0x83
		param[9] = byte(addr >> 16)
		param[10] = byte(addr >> 8)
		param[11] = byte(addr)
	case MemoryType_PA, MemoryType_PE, MemoryType_PI, MemoryType_PQ: // 截取报文发现这两个是一个类型
		param[8] = 0x80
		param[9] = byte(addr >> 16)
		param[10] = byte(addr >> 8)
		param[11] = byte(addr)
	case MemoryType_Q, MemoryType_A: // 截取报文发现这两个是一个类型
		param[8] = 0x82
		param[9] = byte(addr >> 16)
		param[10] = byte(addr >> 8)
		param[11] = byte(addr)
	case MemoryType_T:
		param[8] = 0x1d //地址不需要 * 8
		param[9] = byte(this.Address >> 16)
		param[10] = byte(this.Address >> 8)
		param[11] = byte(this.Address)
	case MemoryType_Z:
		param[8] = 0x1c //地址不需要 * 8
		param[9] = byte(this.Address >> 16)
		param[10] = byte(this.Address >> 8)
		param[11] = byte(this.Address)
	default: //不认识的类型
		return nil
	}
	return
}

//解析一个项
func (this *Item) Paraser(data []byte) (err error) {
	plen := int(binary.BigEndian.Uint16(data[2:]))
	switch this.MemoryType {
	case MemoryType_C: //地址不需要 * 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_DB:
		plen /= 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_E, MemoryType_I: // 截取报文发现这两个是一个类型
		plen /= 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_M, MemoryType_F: // 截取报文发现这两个是一个类型
		plen /= 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_PA, MemoryType_PE, MemoryType_PI, MemoryType_PQ: // 截取报文发现这两个是一个类型
		plen /= 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_Q, MemoryType_A: // 截取报文发现这两个是一个类型
		plen /= 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_T: //地址不需要 * 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	case MemoryType_Z: //地址不需要 * 8
		for _, point := range this.PointArray {
			this.ParserValue(data, point, plen)
		}
	default: //不认识的类型
		return nil
	}
	return
}

//解析项的一个值
func (this *Item) ParserValue(data []byte, point *S7Point, plen int) (err error) {
	t := time.Now().Unix()
	base := 4
	address := int(point.Address - this.Address)
	if address < 0 || address > (len(data)-4) {
		err = errors.New("packet error.")
		logs.Debug(err)
		return
	}
	logs.Debug(address, base, len(data), data, point.Address, this.Address)
	switch point.DataType {
	case DataType_B: //byte
		point.Point.Update(float64(data[address+base]), 0, t)
	case DataType_C: //char
		point.Point.Update(float64(int8(data[address+base])), 0, t)
	case DataType_X: //bit
		point.Point.Update(float64((data[address+base] >> point.Offset)), 0, t)
	case DataType_D: //uint32
		point.Point.Update(float64(binary.BigEndian.Uint32(data[address+base:])), 0, t)
	case DataType_DI: //int32
		point.Point.Update(float64(int32(binary.BigEndian.Uint32(data[address+base:]))), 0, t)
	case DataType_REAL: //float
		val := math.Float32frombits(binary.BigEndian.Uint32(data[address+base:]))
		point.Point.Update(float64(val), 0, t)
	case DataType_W: //uint16
		point.Point.Update(float64(binary.BigEndian.Uint16(data[address+base:])), 0, t)
	case DataType_I: //int16
		point.Point.Update(float64(int16(binary.BigEndian.Uint16(data[address+base:]))), 0, t)
	default:
		err = errors.New("Data type error")
	}
	return
}

//这篇文档来自老外博客
//The S7 protocol is function/command oriented which means a transmission consist of an S7 request and an appropriate reply (with very few exceptions). The number of the parallel transmission and the maximum length of a PDU is negotiated during the connection setup.
//The S7 PDU consists of three main parts:
//Header: contains length information, PDU reference and message type constant
//Parameters: the content and structure greatly varies based on the message and function type of the PDU
//Data: it is an optional field to carry the data if there is any, e.g. memory values, block code, firmware data …etc.

//Protocol ID:[1b] protocol constant always set to 0x32
//Message Type:[1b] the general type of the message (sometimes referred as ROSCTR type)
//0x01-Job Request: request sent by the master (e.g. read/write memory, read/write blocks, start/stop device, setup communication)
//0x02-Ack: simple acknowledgement sent by the slave with no data field (I have never seen it sent by the S300/S400 devices)
//0x03-Ack-Data: acknowledgement with optional data field, contains the reply to a job request
//0x07-Userdata: an extension of the original protocol, the parameter field contains the request/response id, (used for programming/debugging, SZL reads, security functions, time setup, cyclic read..)
//Reserved:[2b] always set to 0x0000 (but probably ignored)
//PDU reference:[2b] generated by the master, incremented with each new transmission, used to link responses to their requests, Little-Endian (note: this is the behaviour of WinCC, Step7, and other Siemens programs, it could probably be randomly generated, the PLC just copies it to the reply)
//Parameter Length:[2b] the length of the parameter field, Big-Endian
//Data Length:[2b] the length of the data field, Big-Endian
//(Error class):[1b] only present in the Ack-Data messages, the possible error constants are listed in the constants.txt
//(Error code):[1b] only present in the Ack-Data messages, the possible error constants are listed in the constants.txt

//PDUreference generated by the master, incremented with each new transmission, used to link responses to their requests, Little-Endian (note: this is the behaviour of WinCC, Step7, and other Siemens programs, it could probably be randomly generated, the PLC just copies it to the reply)
//PUDreference 主站设置,每次连接自动+1,小编码,从站复制这个作为应答
func MakeReauest(PDUreference uint16, items []*Item) (cmd []byte) {
	count := 0
	cmd = make([]byte, 19) //TPKT
	//=================TPKT=================
	cmd[0] = 0x03 //Version
	cmd[1] = 0x00 //Reserved
	cmd[2] = 0x00 //长度高字节(数据包的长度)
	cmd[3] = 0x00 //长度低字节(数据包的长度)

	//ISO 8073/X.224 Connection-Oriented Transport Protocol (面向连接传输协议)
	cmd[4] = 0x02 //长度(COTP长度)
	cmd[5] = 0xf0 //PDU TYPE: DT Data
	cmd[6] = 0x80 //b0~b7: TPDU NUMBER b8:Last data unit:Yes

	//一下是S7 通讯协议
	//S7 Communication
	//Header
	cmd[7] = 0x32                     //Protocol Id(0x32)
	cmd[8] = 0x01                     //ROSCTR:Job (1) 1:Job Request  2:Ack  3:Ack-Data  7 UserData
	cmd[9] = 0x00                     //Redundancy Identification (Reserved) 高字节 //Big-Endian
	cmd[10] = 0x00                    //Redundancy Identification (Reserved) 低字节
	cmd[11] = byte(PDUreference >> 8) //Protocal Data Unit Reference
	cmd[12] = byte(PDUreference)      //Protocal Data Unit Reference
	cmd[15] = 0x00                    //数据长度(高字节)the length of the data field, Big-Endian
	cmd[16] = 0x00                    //参数长度(低字节)
	cmd[17] = 0x04                    //Function: Read var(0x04)

	//每个ITEM 12 个字节, count = 1//使用循环保持后期扩展
	for _, item := range items {
		param := item.MakeRequest()
		if len(param) == 0 {
			continue
		}
		count++
		cmd = append(cmd, param...)
	}
	l := len(cmd)
	cmd[2] = byte(l >> 8) //包长度
	cmd[3] = byte(l)
	paramLen := 2 + count*12      //每个ITEM 12个字节
	cmd[13] = byte(paramLen >> 8) //参数长度(高字节) the length of the parameter field, Big-Endian
	cmd[14] = byte(paramLen)      //参数长度(低字节)
	cmd[18] = byte(count)         //Item count //这里每个组只请求一个ITEM
	if count == 0 {
		return nil
	}
	return
}
