package udp2op

import (
	"encoding/binary"
	"errors"
	"global"
	"math"
	"net"
	"scada/points"
	"strings"
	"time"
	"wj/sock"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

const (
	DefaultValueType    = 4 //R8_TYPE
	DefaultSendInterval = 15
	MaxPointSend        = 1000
)

func Run(conf config.Configer) (err error) {
	const MaxBufLen = 8192
	address := conf.String("destination_address")
	if address == "" {
		err = errors.New("Empty adress for destination")
		return
	}
	isUseServerTime := conf.DefaultBool("use_server_time", false)

	addrs := strings.Split(address, ";")
	conns := make([]*net.UDPConn, len(addrs))
	for i, addr := range addrs {
		conns[i], err = sock.CreateUdpConnect(addr)
		if err != nil {
			logs.Info("Error ", addr, err)
			conns[i] = nil
		}
	}

	go func() {
		length := 14
		num := 0
		buf := make([]byte, MaxBufLen)
		for global.IsRunning() {
			select {
			case <-time.After(time.Millisecond * 200):
				if length > 14 {
					WriteUint16(buf, uint16(length))
					WriteUint16(buf[2:], 255)
					WriteUint16(buf[4:], 0)
					buf[6] = 0xf
					if isUseServerTime {
						buf[7] = 1
					} else {
						buf[7] = 3
					}
					WriteUint16(buf[8:], uint16(num))
					WriteUint32(buf[10:], uint32(points.GetUpdateTime()))
					for _, conn := range conns {
						if conn == nil {
							continue
						}
						conn.Write(buf[:length])
						time.Sleep(time.Millisecond * 40)
					}
					length = 14
					num = 0
				}
			case point := <-points.GetQueuePoint():
				length += WriteUint32(buf[length:], uint32(point.ID))
				length += WriteByte(buf[length:], byte(point.Type))
				switch point.Type {
				case points.AX_TYPE:
					length += WriteUint16(buf[length:], point.Status)
					length += WriteFloat32(buf[length:], float32(point.Value))
				case points.DX_TYPE:
					length += WriteUint16(buf[length:], point.Status)
				case points.I2_TYPE:
					length += WriteUint16(buf[length:], point.Status)
					length += WriteUint16(buf[length:], uint16(point.Value))
				case points.I4_TYPE:
					length += WriteUint16(buf[length:], point.Status)
					length += WriteUint32(buf[length:], uint32(point.Value))
				case points.R8_TYPE:
					length += WriteUint16(buf[length:], point.Status)
					length += WriteFloat64(buf[length:], point.Value)
				default:
					length -= 5
					continue
				}
				num++

				if length > (MaxBufLen - 50) {
					WriteUint16(buf, uint16(length))
					WriteUint16(buf[2:], 255)
					WriteUint16(buf[4:], 0)
					buf[6] = 0xf
					if isUseServerTime {
						buf[7] = 1
					} else {
						buf[7] = 3
					}
					WriteUint16(buf[8:], uint16(num))
					WriteUint32(buf[10:], uint32(points.GetUpdateTime()))
					for _, conn := range conns {
						if conn == nil {
							continue
						}
						conn.Write(buf[:length])
						time.Sleep(time.Millisecond * 40)
					}
					length = 14
					num = 0
				}
			}
		}
	}()
	return
}

func WriteUint32(buf []byte, value uint32) int {
	binary.BigEndian.PutUint32(buf, value)
	return 4
}

func WriteUint16(buf []byte, value uint16) int {
	binary.BigEndian.PutUint16(buf, value)
	return 2
}

func WriteFloat32(buf []byte, value float32) int {
	bits := math.Float32bits(value)
	binary.BigEndian.PutUint32(buf, bits)
	return 4
}

func WriteFloat64(buf []byte, value float64) int {
	bits := math.Float64bits(value)
	binary.BigEndian.PutUint64(buf, bits)
	return 8
}

func WriteByte(buf []byte, value byte) int {
	buf[0] = value
	return 1
}
