package iec104

const (
	M_SP_NA = 1  //单点信息
	M_ME_NA = 9  //测量值, 规一化值
	M_ME_NC = 13 //测量值, 短浮点数 0xd
	M_IT_NA = 15 //累积量 电度数据
	M_SP_TB = 30 //带时标遥信
	M_ME_ND = 21 //不带品质描述的归一化值
	M_IT_TB = 37 //带时标CP56Time2a的累计量

	C_SC_NA         = 45   //0x2d
	C_SC_NA_SELECT  = 0x80 //选择
	C_SC_NA_EXECUTE = 0x00 //执行

	C_IC_NA = 100 //0x64,总召
	C_CI_NA = 101 //0x65, 电度数据召唤

	S_ACK       = 0x1
	STARTDT     = 0x07
	STARTDT_ACT = 0x0b
	STOPDT      = 0x13
	STOPDT_ACT  = 0x23
	TESTFR      = 0x43
	TESTFR_ACT  = 0x83
)

type Head struct {
	start byte   //起始字 0x68
	legth byte   //长度, 最大253
	sn    uint16 //发送序列号
	rn    uint16 //接收序列号
}

type AsduHead struct {
	tid           byte   //类型标识
	vsq           byte   //可变结构限定词
	cot           uint16 //CAUSE OF TRANSMISSION 传输原因
	commonAddress uint16 //COMMON ADDRESS OF ASDU公共地址, 通常RTU地址
}
