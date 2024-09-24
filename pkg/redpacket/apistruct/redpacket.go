package apistruct

// RedPacketType 定义红包类型
type RedPacketType string

const (
	Private   RedPacketType = "private"   // 私聊发送红包
	Luck      RedPacketType = "luck"      // 群聊拼手气红包
	Exclusive RedPacketType = "exclusive" // 群聊用户专属红包
)

// RedPacket 定义红包结构体
type RedPacket struct {
	ReceiveUserID string        `json:"receiveUserId"` // 领取人ID
	Amount        string        `json:"amount"`        // 金额
	Type          RedPacketType `json:"type"`          // 红包类型
	Remark        string        `json:"remark"`        // 备注
	Password      string        `json:"password"`      // 支付密码
	GroupID       string        `json:"groupID"`       // 群ID
	TotalCount    int           `json:"totalCount"`    // 红包个数
	Emoji         string        `json:"emoji"`         // emoji
}
