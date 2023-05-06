package consts

// 服务状态
const (
	Idle = iota
	Working
)

// Conn 类型
const (
	ClientConn = iota
	WorkingConn
)

const (
	HeartBeatType = 100
	HeartBeatCode = 100
	WorkingCode   = 200
)
