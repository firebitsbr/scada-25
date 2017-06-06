package ppi

import "scada/points"

type Group struct {
	Name       string          //点类型
	Size       byte            //数据个数
	Unit       byte            //表示读取数据的单位。为01时，1bit；为02时，1字节；为04时，1字；为06时，双字。
	Area       byte            //存储器类型 01 V存储器; 00 其他
	Type       byte            //表示软器件类型。为04时，S；为05时，SM；为06时，AI；为07时AQ；为1E时，C；为81时，I；为82时，Q；为83时，M；为84时，V；为1F时，T
	Offset     []byte          //偏移量 3个字节
	PointArray []*points.Point //这个组包含的采集点表
}

const (
	S7_COILS        = 1  //01;开关量
	S7_INT16        = 2  //02;有符号短整型
	S7_UINT16       = 3  //03;无符号短整型
	S7_INT32        = 4  //04;有符号整型
	S7_INT32_L      = 5  //05;有符号整型
	S7_UINT32       = 6  //06;无符号整型
	S7_UINT32_L     = 7  //07;无符号整型
	S7_FLOAT        = 8  //08;浮点型
	S7_FLOAT_L      = 9  //09;浮点型
	S7_INT64        = 10 //10;有符号长整型	(暂不支持)
	S7_INT64_L      = 11 //11;无符号长整型	(暂不支持)
	S7_UINT64       = 12 //12;有符号长整型	(暂不支持)
	S7_UINT64_L     = 13 //13;无符号长整型	(暂不支持)
	S7_FLOAT64      = 14 //保留
	S7_FLOAT64_L    = 15 //保留
	S7_FLOAT32_3412 = 16 //保留
)
