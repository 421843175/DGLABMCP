package mcp

import (
	"mygodblab/internal/coyote"
	// 移除未使用的 pulse 导入
)

// Service MCP服务层
type Service struct {
	controller *coyote.Controller
}

// NewService 创建新的Service实例
func NewService(controller *coyote.Controller) *Service {
	return &Service{controller: controller}
}

// SetStrength 设置通道强度
func (s *Service) SetStrength(channel string, strength int) error {
	return s.controller.SetStrength(channel, strength)
}

// SetLimit 设置通道强度上限
func (s *Service) SetLimit(channel string, limit int) error {
	return s.controller.SetLimit(channel, limit)
}

// SetPulse 设置波形
func (s *Service) SetPulse(pulseID string) error {
	return s.controller.SetPulse(pulseID)
}

// GetStatus 获取设备状态
func (s *Service) GetStatus() DeviceStatus {
	// 使用正确的方法名 GetStatus
	channelState := s.controller.GetStatus()

	return DeviceStatus{
		Connected: s.controller.IsConnected(),
		AChannel: ChannelStatus{
			Strength: channelState.AStrength,
			Limit:    channelState.ALimit,
		},
		BChannel: ChannelStatus{
			Strength: channelState.BStrength,
			Limit:    channelState.BLimit,
		},
		CurrentPulse: channelState.CurrentPulse,
		BatteryLevel: channelState.BatteryLevel,
	}
}

// ListPulses 获取可用波形列表
func (s *Service) ListPulses() []PulseInfo {
	// 使用正确的方法名 GetPulseList
	pulses := s.controller.GetPulseList()
	result := make([]PulseInfo, 0, len(pulses))

	for _, p := range pulses {
		result = append(result, PulseInfo{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	return result
}
