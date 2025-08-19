package pulse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// PulseData 波形数据结构
type PulseData struct {
	ID        string   `json:"id" yaml:"id"`
	Name      string   `json:"name" yaml:"name"`
	PulseData []string `json:"pulseData" yaml:"pulse_data"`
}

// Manager 波形管理器
type Manager struct {
	pulses map[string]*PulseData
}

// NewManager 创建波形管理器
func NewManager(configPath string) (*Manager, error) {
	/*等价于
	// 1. 创建一个新的map
	m := make(map[string]*PulseData)

	// 2. 创建Manager结构体
	manager := &Manager{
		pulses: m,
	}
	*/

	//返回一个Manager类型的指针
	/*
		 性能考虑 ：
		- Manager 结构体包含一个 map
		- 如果不用指针，每次传递 Manager 都会复制整个 map
		- 使用指针可以避免不必要的复制，提高性能
	*/
	manager := &Manager{
		pulses: make(map[string]*PulseData),
	}

	err := manager.loadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// NewDefaultManager 创建默认波形管理器
func NewDefaultManager() *Manager {
	manager := &Manager{
		pulses: make(map[string]*PulseData),
	}

	// 添加默认波形
	defaultPulses := []*PulseData{
		{
			ID:   "d6f83af0",
			Name: "呼吸",
			PulseData: []string{
				"0A0A0A0A00000000",
				"0A0A0A0A14141414",
				"0A0A0A0A28282828",
				"0A0A0A0A3C3C3C3C",
				"0A0A0A0A50505050",
				"0A0A0A0A64646464",
				"0A0A0A0A64646464",
				"0A0A0A0A64646464",
				"0A0A0A0A00000000",
				"0A0A0A0A00000000",
				"0A0A0A0A00000000",
				"0A0A0A0A00000000",
			},
		},
		{
			ID:   "7eae1e5f",
			Name: "潮汐",
			PulseData: []string{
				"0A0A0A0A00000000",
				"0D0D0D0D0F0F0F0F",
				"101010101E1E1E1E",
				"1313131332323232",
				"1616161641414141",
				"1A1A1A1A50505050",
				"1D1D1D1D64646464",
				"202020205A5A5A5A",
				"2323232350505050",
				"262626264B4B4B4B",
				"2A2A2A2A41414141",
				"0A0A0A0A00000000",
			},
		},
		{
			ID:   "eea0e4ce",
			Name: "连击",
			PulseData: []string{
				"0A0A0A0A64646464",
				"0A0A0A0A00000000",
				"0A0A0A0A64646464",
				"0A0A0A0A41414141",
				"0A0A0A0A1E1E1E1E",
				"0A0A0A0A00000000",
				"0A0A0A0A00000000",
				"0A0A0A0A00000000",
				"0A0A0A0A64646464",
				"0A0A0A0A00000000",
				"0A0A0A0A64646464",
				"0A0A0A0A00000000",
			},
		},
	}

	for _, pulse := range defaultPulses {
		manager.pulses[pulse.ID] = pulse
	}

	return manager
}

// loadFromFile 从文件加载波形配置
func (m *Manager) loadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// 尝试JSON格式
	var jsonPulses []*PulseData
	if err := json.Unmarshal(data, &jsonPulses); err == nil {
		for _, pulse := range jsonPulses {
			m.pulses[pulse.ID] = pulse
		}
		return nil
	}

	// 尝试YAML格式
	var yamlPulses []*PulseData
	if err := yaml.Unmarshal(data, &yamlPulses); err == nil {
		for _, pulse := range yamlPulses {
			m.pulses[pulse.ID] = pulse
		}
		return nil
	}

	return fmt.Errorf("无法解析波形配置文件")
}

// GetPulse 获取指定ID的波形
func (m *Manager) GetPulse(id string) (*PulseData, error) {
	pulse, exists := m.pulses[id]
	if !exists {
		return nil, fmt.Errorf("波形不存在: %s", id)
	}
	return pulse, nil
}

// ListPulses 列出所有波形
func (m *Manager) ListPulses() []*PulseData {
	var pulses []*PulseData
	for _, pulse := range m.pulses {
		pulses = append(pulses, pulse)
	}
	return pulses
}

// AddPulse 添加新波形
func (m *Manager) AddPulse(pulse *PulseData) {
	m.pulses[pulse.ID] = pulse
}
