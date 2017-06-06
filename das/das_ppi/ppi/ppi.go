package ppi

import (
	"errors"
	"fmt"
	"io"
	"opapi4/opevent"
	"scada/points"
	"sort"
	"strings"

	"github.com/astaxie/beego/logs"
)

type PPI struct {
	pointPath  string
	da         byte     //目标地址，占1字节，指PLC在PPI上地址，一台PLC时，一般为02，多台PLC时，则各有各的地址；
	sa         byte     //源地址，占1字节，指计算机在PPI上地址，一般为00；
	groupArray []*Group //分组
}

func New(pointList string) (das points.Daser, err error) {
	d := new(PPI)
	d.pointPath = pointList
	err = d.LoadPointList()
	das = d
	return
}

func (this *PPI) Work(conn interface{}, control opevent.Controler) (err error) {
	return
}

//加载点表
func (this *PPI) LoadPointList() (err error) {
	this.groupArray = make([]*Group, 0)
	points.PFuncFormatAddress = FormatAddress
	pointArray, err := points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		return
	}

	//对数据类型进行排序
	sort.Slice(pointArray, func(i, j int) bool {
		if pointArray[i].PointType < pointArray[j].PointType {
			return true
		}
		return false
	})
	points.Append(pointArray...)

	//数据分组
	for _, point := range pointArray {
		this.AddPointToGroup(point)
	}

	return
}

//转换modbus对应的数据类型
func GetType(pt string) byte {
	modbusType := byte(0)
	switch pt {
	case "COILS":
		modbusType = S7_COILS
	case "INT8", "INT16":
		modbusType = S7_INT16
	case "UINT16":
		modbusType = S7_UINT16
	case "INT32":
		modbusType = S7_INT32
	case "INT32_L":
		modbusType = S7_INT32_L
	case "UINT32":
		modbusType = S7_UINT32
	case "UINT32_L":
		modbusType = S7_UINT32_L
	case "FLOAT":
		modbusType = S7_FLOAT
	case "FLOAT_L":
		modbusType = S7_FLOAT_L
	case "FLOAT_3412":
		modbusType = S7_FLOAT32_3412
	case "INT64":
		modbusType = S7_INT64
	case "INT64_L":
		modbusType = S7_INT64_L
	case "FLOAT64":
		modbusType = S7_FLOAT64
	case "FLOAT64_L":
		modbusType = S7_FLOAT64_L
	default:
		modbusType = S7_INT16
	}
	return modbusType
}

//解析地址信息
func FormatAddress(point *points.Point, address string) (err error) {
	ss := strings.Split(address, ";")
	if len(ss) > 0 {
		if len(ss[0]) == 0 {
			err = errors.New("No address")
			return
		}
		if len(ss) == 1 {
			ss[0] = strings.TrimSpace(ss[0])

		} else if len(ss) == 2 {
			ss[1] = strings.TrimSpace(ss[1])
		}
	}
	err = errors.New("No address")
	return
}

//把点添加到组中
//加载全部点表后排序,按照顺序加入点组,当一个组中的最大下标比第一个下标大于100时(四字节为50)另起一组
func (this *PPI) AddPointToGroup(point *points.Point) (err error) {

	return
}

//报文结构  SD  LE LER  SD  DA SA  FC  DASP SSAP  DU  FCS  ED
//SD:(Start Delimiter)开始定界符，占1字节，为68H
//LE:（Length）报文数据长度，占1字节，标明报文以字节计，从DA到DU的长度；
//LER:（Repeated Length）重复数据长度，同LE
//SD: (Start Delimiter)开始定界符(68H)
//DA:（DestinationAddress）目标地址，占1字节，指PLC在PPI上地址，一台PLC时，一般为02，多台PLC时，则各有各的地址；
//SA:（Source Address）源地址，占1字节，指计算机在PPI上地址，一般为00；
//FC:（Function Code）功能码，占1字节，6CH一般为读数据，7CH一般为写数据
//DSAP:（Destination Service Access Point）目的服务存取点，占多个字节
//SSAP:（Source Service Access Point）源服务存取点，占多个字节
//DU:（Data Unit）数据单元，占多个字节
//FCS:（Frame CheckSequence）占1字节，从DA到DU之间的校验和的256余数；
//ED:（End Delimiter）结束分界符，占1字节，为16H
func (this *PPI) MakeRequest(group *Group) (buf []byte) {
	buf = make([]byte, 33)
	copy(buf, []byte{0x68, 0x1b, 0x1b, 0x68}) //固定结构
	buf[4] = this.da                          //目标地址
	buf[5] = this.sa                          //源地址
	buf[6] = 0x6c                             //功能码 6CH一般为读数据，7CH一般为写数据
	copy(buf[7:], []byte{0x32, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x0E, 0x00, 0x00, 0x04, 0x01, 0x12,
		0x0A, 0x10}) //固定结构,不知道什么意思
	buf[22] = group.Unit           //表示读取数据的单位。为01时，1bit；为02时，1字节；为04时，1字；为06时，双字。
	buf[23] = 0                    //恒0。
	buf[24] = group.Size           //字节24，表示数据个数。01，表示一次读一个数据。如为读字节，最多可读208个字节，即可设为DEH。
	buf[25] = 0                    //恒0。
	buf[26] = 0                    //表示软器件类型。为01时，V存储器；为00时，其它。
	buf[27] = group.Type           //表示软器件类型。为04时，S；为05时，SM；为06时，AI；为07时AQ；为1E时，C；为81时，I；为82时，Q；为83时，M；为84时，V；为1F时，T。
	copy(buf[28:31], group.Offset) //偏移量
	buf[31] = FCS(buf[4:31])       //从DA到DU之间的校验和的256余数；
	buf[32] = 0x16                 //结束分界符，占1字节，为16H
	return
}

func FCS(data []byte) (fcs byte) {
	fcs = 0
	for _, v := range data {
		fcs += v
	}
	return
}

func Send(data []byte, conn io.Writer) (err error) {
	if conn == nil {
		err = errors.New("Empty connection.")
		return
	}
	logs.Debug("S: ", fmt.Sprintf("% 02X\n", data))
	_, err = conn.Write(data)
	return
}
