package s7_300

import (
	"errors"
	"scada/points"
	"strconv"
	"strings"
)

//Address Syntax
//Input, Output, Peripheral, Flag Memory Types
//<memory type><S7 data type><address>
//<memory type><S7 data type><address><.bit>
//<memory type><S7 data type><address><.string length>*
//<memory type><S7 data type><address><[row][>col]>

//DB Memory Type
//DB<num>,<S7 data type><address>
//DB<num>,<S7 data type><address><.bit>
//DB<num>,<S7 data type><address><.string length>*
//DB<num>,<S7 data type><address><[row][col]>

const (
	MemoryType_I  = 1 + iota //read/write
	MemoryType_E             //read/write
	MemoryType_Q             //read/write
	MemoryType_A             //read/write
	MemoryType_PI            //read only
	MemoryType_PE            //read only
	MemoryType_PQ            //read/write
	MemoryType_PA            //read/write
	MemoryType_M             //read/write
	MemoryType_F             //read/write
	MemoryType_DB            //read/write
	MemoryType_T             //read/write
	MemoryType_C             //read/write
	MemoryType_Z             //read/write
)

const (
	//以下代码必须打字节类型在前
	DataType_STRING = 1 + iota //string, STRING0.n-STRING65532.n  .n is string length.  0<n<= 210.
	DataType_REAL              //float
	DataType_DI                //int32
	DataType_D                 //uint32
	DataType_I                 //int16
	DataType_W                 //uint16
	DataType_C                 //int8, char
	DataType_B                 //uint8,byte
	DataType_X                 //bit,boolean
)

type S7Point struct {
	Point      *points.Point //指向关联的测点
	MemoryType byte
	DataType   byte
	DBBlock    uint16 //only for DB, DB1.B0.64
	Address    uint16
	Offset     byte
}

//把地址解析成s7point
func Parser2S7Point(address string) (point *S7Point, err error) {
	address = strings.TrimSpace(address)
	address = strings.ToUpper(address)
	point = new(S7Point)
	s := GetPrefix(address)
	if s == "" {
		err = errors.New("Error address")
		return
	}

	//不能用switch, 因为还包含类型 IB0.7
	if strings.HasPrefix(s, "I") {
		point.MemoryType = MemoryType_I
		err = ParserDataTypeAddress(address[1:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "E") {
		point.MemoryType = MemoryType_E
		err = ParserDataTypeAddress(address[1:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "Q") {
		point.MemoryType = MemoryType_Q
		err = ParserDataTypeAddress(address[1:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "A") {
		point.MemoryType = MemoryType_A
		err = ParserDataTypeAddress(address[1:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "PI") {
		point.MemoryType = MemoryType_PI
		err = ParserDataTypeAddress(address[2:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "PE") {
		point.MemoryType = MemoryType_PE
		err = ParserDataTypeAddress(address[2:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "PA") {
		point.MemoryType = MemoryType_PA
		err = ParserDataTypeAddress(address[2:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "M") {
		point.MemoryType = MemoryType_M
		err = ParserDataTypeAddress(address[1:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "F") {
		point.MemoryType = MemoryType_F
		err = ParserDataTypeAddress(address[1:], point) //去掉MemoryType 标志
	} else if strings.HasPrefix(s, "DB") {
		point.MemoryType = MemoryType_DB
		//获取DB Block
		ss := strings.SplitN(address, ".", 2)
		if len(ss) < 2 {
			err = errors.New("Error address.")
			return
		}
		db := 0
		db, err = strconv.Atoi(ss[0][2:])
		if err != nil {
			return nil, err
		}
		point.DBBlock = uint16(db)
		err = ParserDataTypeAddress(ss[1], point)
	} else if strings.HasPrefix(s, "T") { //T0-T65535
		point.MemoryType = MemoryType_T
		point.DataType = DataType_DI
		err = ParserAddress(address[1:], point)
	} else if strings.HasPrefix(s, "C") { //C0-C65535
		point.MemoryType = MemoryType_C
		point.DataType = DataType_I
		err = ParserAddress(address[1:], point)
	} else if strings.HasPrefix(s, "Z") { //Z0-Z65535
		point.DataType = DataType_I
		point.MemoryType = MemoryType_Z
		err = ParserAddress(address[1:], point)
	} else {
		err = errors.New("Error memory type.")
		return
	}
	return
}

//解析地址填充到对应的结构体中 Only for T,C,Z
func ParserAddress(address string, point *S7Point) (err error) {
	addr, err := strconv.Atoi(address)
	point.Address = uint16(addr)
	return
}

//解析地址填充到对应的结构体中
func ParserDataTypeAddress(address string, point *S7Point) (err error) {
	ss := strings.Split(address, ".")
	if len(ss) == 0 || len(ss[0]) == 0 {
		err = errors.New("Address error")
		return
	}

	//获取类型和地址号
	addr := 0
	s := GetPrefix(ss[0])
	switch s {
	case "": //默认为byte
		point.DataType = DataType_B
		addr, err = strconv.Atoi(ss[0])
		point.Address = uint16(addr)
	case "X": //bit
		point.DataType = DataType_X
		addr, err = strconv.Atoi(ss[0][1:]) //移出类型标志
		point.Address = uint16(addr)
	case "B": //"BYTE"
		point.DataType = DataType_B
		addr, err = strconv.Atoi(ss[0][1:]) //移出类型标志
		point.Address = uint16(addr)
	case "C": // "CHAR"
		point.DataType = DataType_C
		addr, err = strconv.Atoi(ss[0][1:]) //移出类型标志
		point.Address = uint16(addr)
	case "W": // "WORD"
		point.DataType = DataType_W
		addr, err = strconv.Atoi(ss[0][1:]) //移出类型标志
		point.Address = uint16(addr)
	case "I": // "INT"
		point.DataType = DataType_I
		addr, err = strconv.Atoi(ss[0][1:]) //移出类型标志
		point.Address = uint16(addr)
	case "D": // "DWORD"
		point.DataType = DataType_D
		addr, err = strconv.Atoi(ss[0][1:]) //移出类型标志
		point.Address = uint16(addr)
	case "DB":
		point.DataType = DataType_B
		addr, err = strconv.Atoi(ss[0][2:]) //移出类型标志
		point.Address = uint16(addr)
	case "DBB":
		point.DataType = DataType_B
		addr, err = strconv.Atoi(ss[0][3:]) //移出类型标志
		point.Address = uint16(addr)
	case "DBW":
		point.DataType = DataType_W
		addr, err = strconv.Atoi(ss[0][3:]) //移出类型标志
		point.Address = uint16(addr)
	case "DBD":
		point.DataType = DataType_D
		addr, err = strconv.Atoi(ss[0][3:]) //移出类型标志
		point.Address = uint16(addr)
	case "DI": // "DINT"
		point.DataType = DataType_DI
		addr, err = strconv.Atoi(ss[0][2:]) //移出类型标志
		point.Address = uint16(addr)
	case "REAL":
		point.DataType = DataType_REAL
		addr, err = strconv.Atoi(ss[0][4:]) //移出类型标志
		point.Address = uint16(addr)
	case "STRING":
		point.DataType = DataType_STRING
		addr, err = strconv.Atoi(ss[0][6:]) //移出类型标志
		point.Address = uint16(addr)
	default:
		err = errors.New("Address error")
		return
	}
	//获取偏移量
	if len(ss) == 2 {
		off := 0
		off, err = strconv.Atoi(ss[1])
		point.Offset = byte(off)
	}
	return
}

func GetPrefix(s string) string {
	for i := range s {
		if s[i] <= 'Z' && s[i] >= 'A' {
			continue
		}
		return s[:i]
	}
	return ""
}

func GetMemoryType(t byte) string {
	switch t {
	case MemoryType_I:
		return "I"
	case MemoryType_E:
		return "E"
	case MemoryType_Q:
		return "Q"
	case MemoryType_A:
		return "A"
	case MemoryType_PI:
		return "PI"
	case MemoryType_PE:
		return "PE"
	case MemoryType_PQ:
		return "PQ"
	case MemoryType_PA:
		return "PA"
	case MemoryType_M:
		return "M"
	case MemoryType_F:
		return "F"
	case MemoryType_DB:
		return "DB"
	case MemoryType_T:
		return "T"
	case MemoryType_C:
		return "C"
	case MemoryType_Z:
		return "Z"
	}
	return ""
}

func GetDataType(t byte) string {
	switch t {
	case DataType_X: //= 1 + iota //bit,boolean
		return "boolean"
	case DataType_B: //uint8,byte
		return "byte"
	case DataType_C: //int8, char
		return "char"
	case DataType_W: //uint16
		return "uint16"
	case DataType_I: //int16
		return "int16"
	case DataType_D: //uint32
		return "uint32"
	case DataType_DI: //int32
		return "int32"
	case DataType_REAL: //float
		return "float"
	case DataType_STRING: //string, STRING0.n-STRING65532.n  .n is string length.  0<n<= 210.
		return "string"
	}
	return ""
}
