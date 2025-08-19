package mcp

// MCPRequest 定义了MCP请求的结构
type MCPRequest struct {
	Action  string      `json:"action"`  // 操作类型
	Payload interface{} `json:"payload"` // 请求参数
}

// MCPResponse 定义了MCP响应的结构
type MCPResponse struct {
	Success bool        `json:"success"` // 操作是否成功
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// SetStrengthRequest 设置强度请求
type SetStrengthRequest struct {
	Channel  string `json:"channel"`  // 通道（A或B）
	Strength int    `json:"strength"` // 强度值（0-200）
}

// SetLimitRequest 设置上限请求
type SetLimitRequest struct {
	Channel string `json:"channel"` // 通道（A或B）
	Limit   int    `json:"limit"`   // 上限值（0-200）
}

// SetPulseRequest 设置波形请求
type SetPulseRequest struct {
	PulseID string `json:"pulse_id"` // 波形ID
}

// ChannelStatus 通道状态
type ChannelStatus struct {
	Strength int `json:"strength"` // 当前强度
	Limit    int `json:"limit"`    // 强度上限
}

// DeviceStatus 设备状态
type DeviceStatus struct {
	Connected    bool          `json:"connected"`     // 连接状态
	AChannel     ChannelStatus `json:"a_channel"`     // A通道状态
	BChannel     ChannelStatus `json:"b_channel"`     // B通道状态
	CurrentPulse string        `json:"current_pulse"` // 当前波形ID
	BatteryLevel int           `json:"battery_level"` // 电量百分比
}

// PulseInfo 波形信息
type PulseInfo struct {
	ID   string `json:"id"`   // 波形ID
	Name string `json:"name"` // 波形名称
}
