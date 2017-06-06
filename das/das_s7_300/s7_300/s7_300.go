package s7_300

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"opapi4/opevent/opeventer"
	"scada/points"
	"sort"
	"strings"
	"time"
	"wj/sock"

	"github.com/astaxie/beego/config"

	"github.com/astaxie/beego/logs"
)

type S7 struct {
	itemArray []*Item
	pointPath string
	pudr      uint16
}

func New(conf config.Configer) (das points.Daser, err error) {
	d := new(S7)

	d.pointPath = conf.DefaultString("service_name", points.GetAppName())
	err = d.LoadPointList()
	das = d
	return
}

func (this *S7) Check() {
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

//加载点表
func (this *S7) LoadPointList() (err error) {
	points.PFuncFormatAddress = FormatAddress
	pointArray, err := points.GetPointFromDB(this.pointPath, false)
	if err != nil {
		logs.Error(err)
		return
	}

	//排序
	sort.Slice(pointArray, func(i, j int) bool {
		point1, ok1 := pointArray[i].VoidPtr.(*S7Point)
		point2, ok2 := pointArray[i].VoidPtr.(*S7Point)
		if ok1 && ok2 {
			if point1.MemoryType < point2.MemoryType { //比较内存类型
				return true
			} else if point1.MemoryType == point2.MemoryType {
				if point1.Address < point2.Address { //比较地址
					return true
				} else if point1.Address == point2.Address {
					if point1.Offset < point2.Offset { //比较偏移量
						return true
					}
				}
			}
		}
		return false
	})

	//Item 分配到对应的项中
	this.distribute2Item(pointArray)

	points.Append(pointArray...)
	return err
}

//工作主函数
func (this *S7) Work(conn interface{}, control opeventer.Controler) (err error) {
	this.Check()
	if tcpconn, ok := conn.(io.ReadWriter); ok {
		var data []byte
		//设置链路
		cmd := this.initializeCommand()
		logs.Debug("S: % 02x", cmd)
		_, err = tcpconn.Write(cmd)
		if err != nil {
			return
		}
		data, err = RecvCommand(tcpconn)
		if err != nil {
			return
		}
		logs.Debug("R: % 02x", data)

		//Setup s7-300/400
		cmd = this.setupCommand()
		logs.Debug("S: % 02x", cmd)
		_, err = tcpconn.Write(cmd)
		if err != nil {
			return
		}
		data, err = RecvCommand(tcpconn)
		if err != nil {
			return
		}
		logs.Debug("R: % 02x", data)

		for {
			this.Check()
			l := len(this.itemArray)
			for i := 0; i < l; i++ {
				err = this.ReadData(tcpconn, this.pudr, this.itemArray[i:i+1])
				if err != nil {
					return
				}
			}
			time.Sleep(time.Millisecond * 1000)
		}
	}
	return
}

/*初始化西门子的数据请求命令
 */
func (this *S7) initializeCommand() (cmd []byte) {
	cmd = []byte{0x03, 0x00, 0x00, 0x16, 0x11, 0xE0, 0x00, 0x00, 0x00, 0x01,
		0x00, 0xC0, 0x01, 0x09, 0xC1, 0x02, 0x4B, 0x54, 0xC2, 0x02, 0x03, 0x02}
	return
}

/*西门子的数据请求
 */
func (this *S7) setupCommand() (cmd []byte) {
	cmd = []byte{0x03, 0x00, 0x00, 0x19, 0x02, 0xF0, 0x80, 0x32, 0x01, 0x00, 0x00,
		0x00, 0x02, 0x00, 0x08, 0x00, 0x00, 0xF0, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0xF0}
	return
}

//把测点分配到组中
func (this *S7) distribute2Item(points []*points.Point) {
	this.itemArray = make([]*Item, 1)
	item := new(Item)
	this.itemArray[0] = item
	for _, point := range points {
		if p, ok := point.VoidPtr.(*S7Point); ok {
			//log.Println(p)
			err := item.AddPoint(p)
			if err != nil {
				item = new(Item)
				item.AddPoint(p)
				this.itemArray = append(this.itemArray, item)
			}
		}
	}
}

func (this *S7) ReadData(conn io.ReadWriter, pudr uint16, items []*Item) (err error) {
	//log.Println("Item number:", len(items))
	cmd := MakeReauest(pudr, items)

	//log.Printf("%d % 02x", len(cmd), cmd)
	_, err = conn.Write(cmd)
	logs.Debug("S: % 02x", cmd)
	if err != nil {
		logs.Info(err)
		return
	}
	data, err := RecvCommand(conn)
	if err != nil {
		logs.Info(err)
		return
	}
	logs.Debug("R: % 02x", data)
	verr := this.VerifyData(data, pudr)
	if verr != nil {
		//log.Println(err)
		logs.Info(err)
		return
	}

	//从数据开始部分解析函数
	perr := this.ParserData(data[20:], items)
	if perr != nil {
		logs.Info(perr)
	}
	return
}

//校验数据是否正确
func (this *S7) VerifyData(data []byte, pudr uint16) (err error) {
	length := len(data)
	if length < 25 {
		return errors.New("Ack packet length error.")
	}

	if data[0] != 0x03 {
		return errors.New("The version of tpkt is error")
	}

	if data[8] != 0x03 { //ROSCTR:Job (1) 1:Job Request  2:Ack  3:Ack-Data  7 UserData
		return errors.New("Ack packet is not Ack-Data")
	}

	if data[19] != 0x04 { //Function: Read var(0x04)
		return errors.New("Function code is not Read var")
	}

	packetLength := binary.BigEndian.Uint16(data[2:])
	if int(packetLength) != length {
		return errors.New(fmt.Sprintf("packet length error. Packet length=%d, len=%d", packetLength, length))
	}

	//检查pudr 是否一致
	ppudr := binary.BigEndian.Uint16(data[11:])
	if pudr != ppudr {
		return errors.New("Pudr error")
	}

	return
}

//解析数据
func (this *S7) ParserData(data []byte, items []*Item) (err error) {
	//查看请求的和发送的个数是否一致
	nItem := len(items)
	if nItem != int(data[0]) {
		err = errors.New("The number of items is not match." + fmt.Sprint(nItem, data[0]))
		return
	}

	posBegin := 1
	for i := 0; i < nItem; i++ {
		if data[posBegin] != 0xff {
			continue
		}
		items[i].Paraser(data[posBegin:])
	}

	return
}

//格式化地址
func FormatAddress(point *points.Point, address string) (err error) {
	address = strings.TrimSuffix(address, ";")
	p, err := Parser2S7Point(address)
	if err != nil {
		return
	}
	p.Point = point //互相关注
	point.VoidPtr = p
	point.Extend = address
	return
}

func RecvCommand(conn io.ReadWriter) (data []byte, err error) {
	data = make([]byte, 4)
	n, err := sock.ReadTimeout(conn, data[:4], 30*1000)
	if err != nil {
		return
	}
	plen := binary.BigEndian.Uint16(data[2:])
	if plen < 5 {
		err = errors.New("Packet error.")
		return
	}
	plen -= 4
	buf := make([]byte, plen)
	n, err = sock.ReadTimeout(conn, buf, 30*1000)
	if err != nil {
		return
	}

	if uint16(n) != plen {
		err = errors.New("The length of packet error")
	}
	data = append(data, buf...)
	return data, err
}
