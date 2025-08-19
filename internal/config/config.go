package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Bluetooth BluetoothConfig `yaml:"bluetooth"`
	Channels  ChannelConfig   `yaml:"channels"`
	Pulses    PulseConfig     `yaml:"pulses"`
}

// BluetoothConfig 蓝牙配置
type BluetoothConfig struct {
	ScanTimeout int      `yaml:"scan_timeout"` // 扫描超时时间(秒)
	DeviceNames []string `yaml:"device_names"` // 目标设备名称
	ServiceUUID string   `yaml:"service_uuid"` // 服务UUID
	WriteUUID   string   `yaml:"write_uuid"`   // 写特性UUID
	NotifyUUID  string   `yaml:"notify_uuid"`  // 通知特性UUID
	BatteryUUID string   `yaml:"battery_uuid"` // 电量特性UUID
}

// ChannelConfig 通道配置
type ChannelConfig struct {
	AChannel ChannelSettings `yaml:"a_channel"`
	BChannel ChannelSettings `yaml:"b_channel"`
}

// ChannelSettings 单个通道设置
type ChannelSettings struct {
	Enabled         bool `yaml:"enabled"`          // 是否启用
	MaxStrength     int  `yaml:"max_strength"`     // 最大强度
	DefaultStrength int  `yaml:"default_strength"` // 默认强度
}

// PulseConfig 波形配置
type PulseConfig struct {
	ConfigPath     string `yaml:"config_path"`     // 波形配置文件路径
	DefaultPulse   string `yaml:"default_pulse"`   // 默认波形ID
	UpdateInterval int    `yaml:"update_interval"` // 更新间隔(ms)
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Bluetooth: BluetoothConfig{
			ScanTimeout: 30,
			DeviceNames: []string{"47L121000", "47L120100"},
			ServiceUUID: "0000180C-0000-1000-8000-00805f9b34fb",
			WriteUUID:   "0000150A-0000-1000-8000-00805f9b34fb",
			NotifyUUID:  "0000150B-0000-1000-8000-00805f9b34fb",
			BatteryUUID: "00001500-0000-1000-8000-00805f9b34fb",
		},
		Channels: ChannelConfig{
			AChannel: ChannelSettings{
				Enabled:         true,
				MaxStrength:     100,
				DefaultStrength: 0,
			},
			BChannel: ChannelSettings{
				Enabled:         false,
				MaxStrength:     100,
				DefaultStrength: 0,
			},
		},
		Pulses: PulseConfig{
			ConfigPath:     "pulses.yaml",
			DefaultPulse:   "d6f83af0",
			UpdateInterval: 100,
		},
	}
}
