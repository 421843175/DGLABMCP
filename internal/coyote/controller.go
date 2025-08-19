package coyote

import (
	"fmt"
	"log"
	"sync"
	"time"

	"mygodblab/internal/bluetooth" //蓝牙通信包
	"mygodblab/internal/config"    //配置管理包
	"mygodblab/internal/protocol"  //协议包
	"mygodblab/internal/pulse"     //波形管理包
)

// Controller 郊狼设备控制器
// Controller 郊狼设备控制器结构体，管理设备的所有功能
type Controller struct {
	config       *config.Config              // 应用配置信息，包含蓝牙和通道设置
	btAdapter    *bluetooth.BluetoothAdapter // 蓝牙适配器，处理与设备的通信
	pulseManager *pulse.Manager              // 波形管理器，管理各种电击波形模式
	channelState *ChannelState               // 通道状态，记录A/B通道的当前状态
	sequence     byte                        // 指令序列号(0-15)，用于标识每个指令
	mu           sync.RWMutex                // 读写互斥锁，保护并发访问
}

// ChannelState 通道状态
type ChannelState struct {
	AStrength    int    // A通道当前强度
	BStrength    int    // B通道当前强度
	ALimit       int    // A通道强度上限
	BLimit       int    // B通道强度上限
	CurrentPulse string // 当前波形ID
	BatteryLevel int    // 电量百分比
}

// NewController 创建新的控制器实例
func NewController(cfg *config.Config) (*Controller, error) {
	pulseManager, err := pulse.NewManager(cfg.Pulses.ConfigPath)
	if err != nil {
		log.Printf("加载波形配置失败，使用默认配置: %v", err)
		pulseManager = pulse.NewDefaultManager()
	}

	// 创建蓝牙适配器
	btAdapter := bluetooth.NewBluetoothAdapter()

	// 启用蓝牙
	err = btAdapter.Enable()
	if err != nil {
		return nil, fmt.Errorf("启用蓝牙失败: %w", err)
	}

	return &Controller{
		config:       cfg,          // 将传入的配置对象赋值给config字段，包含了所有的应用程序配置信息
		btAdapter:    btAdapter,    // 将传入的蓝牙适配器对象赋值给btAdapter字段，用于处理蓝牙通信
		pulseManager: pulseManager, // 将传入的脉冲管理器对象赋值给pulseManager字段，用于管理波形数据
		channelState: &ChannelState{ // 创建并初始化一个新的ChannelState结构体指针
			AStrength:    cfg.Channels.AChannel.DefaultStrength, // 从配置中获取A通道的默认强度值
			BStrength:    cfg.Channels.BChannel.DefaultStrength, // 从配置中获取B通道的默认强度值
			ALimit:       cfg.Channels.AChannel.MaxStrength,     // 从配置中获取A通道的最大强度限制
			BLimit:       cfg.Channels.BChannel.MaxStrength,     // 从配置中获取B通道的最大强度限制
			CurrentPulse: cfg.Pulses.DefaultPulse,               // 从配置中获取默认的脉冲波形名称
			BatteryLevel: 0,                                     // 初始化电池电量为0，后续将通过蓝牙通信获取实际电量
		},
		sequence: 1, // 初始化指令序列号为1，用于DG-LAB协议的命令同步
	}, nil
}

// ScanAndConnect 扫描并连接到郊狼设备
func (c *Controller) ScanAndConnect(timeout time.Duration) error {
	//调用bluetooth.BluetoothAdapter的ScanAndConnect
	return c.btAdapter.ScanAndConnect(timeout, c.config.Bluetooth.DeviceNames)
}

// SetStrength 设置通道强度
func (c *Controller) SetStrength(channel string, strength int) error {
	if !c.btAdapter.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 验证强度值
	strength = int(protocol.ValidateStrength(strength))

	// 检查上限
	var limit int
	switch channel {
	case "A", "a":
		limit = c.channelState.ALimit
		if strength > limit {
			strength = limit
			log.Printf("强度超过上限，调整为: %d", strength)
		}
		c.channelState.AStrength = strength
	case "B", "b":
		limit = c.channelState.BLimit
		if strength > limit {
			strength = limit
			log.Printf("强度超过上限，调整为: %d", strength)
		}
		c.channelState.BStrength = strength
	default:
		return fmt.Errorf("无效的通道: %s", channel)
	}

	// 构建并发送B0指令
	cmd := c.buildB0Command()
	switch channel {
	case "A", "a":
		cmd.AMode = protocol.StrengthModeAbsolute
		cmd.AStrength = byte(strength)
	case "B", "b":
		cmd.BMode = protocol.StrengthModeAbsolute
		cmd.BStrength = byte(strength)
	}

	log.Printf("设置%s通道强度为: %d", channel, strength)
	return c.sendCommand(cmd)
}

// sendCommand 发送命令到设备
func (c *Controller) sendCommand(cmd *protocol.B0Command) error {
	data := cmd.ToBytes()
	/*
		protocol.B0Command
		{Sequence: 1,
		AMode: StrengthModeAbsolute (3),
		BMode: StrengthModeNoChange (0),
		AStrength: 20,
		BStrength: 0,
		AWaveData: [4]mygodblab/internal/protocol.WaveData
		[4]protocol.WaveData [{Frequency: 10, Strength: 20},{Frequency: 10, Strength: 20},{Frequency: 10, Strength: 20},{Frequency: 10, Strength: 20}]
	*/
	err := c.btAdapter.WriteCharacteristic(data)
	if err != nil {
		return fmt.Errorf("发送命令失败: %w", err)
	}
	return nil
}

// PrintStatus 打印当前状态
func (c *Controller) PrintStatus() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Println("\n=== 设备状态 ===")
	fmt.Printf("连接状态: %v\n", c.btAdapter.IsConnected())
	fmt.Printf("A通道强度: %d/%d\n", c.channelState.AStrength, c.channelState.ALimit)
	fmt.Printf("B通道强度: %d/%d\n", c.channelState.BStrength, c.channelState.BLimit)
	fmt.Printf("当前波形: %s\n", c.channelState.CurrentPulse)
	if c.channelState.BatteryLevel > 0 {
		fmt.Printf("电量: %d%%\n", c.channelState.BatteryLevel)
	}
	fmt.Println()
}

// Close 关闭控制器
func (c *Controller) Close() error {
	if c.btAdapter != nil {
		return c.btAdapter.Disconnect()
	}
	return nil
}

// SetLimit 设置通道强度上限
func (c *Controller) SetLimit(channel string, limit int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	limit = int(protocol.ValidateStrength(limit))

	switch channel {
	case "A", "a":
		c.channelState.ALimit = limit
		// 如果当前强度超过新上限，调整当前强度
		if c.channelState.AStrength > limit {
			c.channelState.AStrength = limit
			log.Printf("A通道当前强度超过新上限，调整为: %d", limit)
		}
	case "B", "b":
		c.channelState.BLimit = limit
		if c.channelState.BStrength > limit {
			c.channelState.BStrength = limit
			log.Printf("B通道当前强度超过新上限，调整为: %d", limit)
		}
	default:
		return fmt.Errorf("无效的通道: %s", channel)
	}

	log.Printf("%s通道强度上限设置为: %d", channel, limit)
	return nil
}

// SetPulse 设置波形
func (c *Controller) SetPulse(pulseID string) error {
	pulseData, err := c.pulseManager.GetPulse(pulseID)
	if err != nil {
		return fmt.Errorf("获取波形失败: %w", err)
	}

	c.mu.Lock()
	c.channelState.CurrentPulse = pulseID
	c.mu.Unlock()

	log.Printf("切换到波形: %s (%s)", pulseData.Name, pulseID)
	return nil
}

// buildB0Command 构建基础B0指令 - 用于创建发送给设备的B0控制指令
// B0 指令写入通道强度变化和通道波形数据,
func (c *Controller) buildB0Command() *protocol.B0Command {
	// 从脉冲管理器获取当前选择的波形数据
	pulseData, _ := c.pulseManager.GetPulse(c.channelState.CurrentPulse)

	// 创建新的B0指令对象，设置序列号和初始模式
	cmd := &protocol.B0Command{
		Sequence: c.sequence,                    // 设置指令序列号
		AMode:    protocol.StrengthModeNoChange, // A通道强度模式设为不变
		BMode:    protocol.StrengthModeNoChange, // B通道强度模式设为不变
	}

	// 设置波形数据 - 检查是否有可用的波形数据
	if pulseData != nil && len(pulseData.PulseData) > 0 {
		// 使用第一组波形数据 - 获取十六进制格式的波形数据
		hexData := pulseData.PulseData[0]
		// 将十六进制数据转换为A通道的波形数据
		if aWaves, err := protocol.WaveDataFromHex(hexData); err == nil {
			cmd.AWaveData = aWaves
		}
		// 将十六进制数据转换为B通道的波形数据
		if bWaves, err := protocol.WaveDataFromHex(hexData); err == nil {
			cmd.BWaveData = bWaves
		}
	} else {
		// 如果没有可用波形，使用默认波形数据：频率10Hz，强度50%
		defaultWave := protocol.WaveData{Frequency: 10, Strength: 50}
		// 为A和B通道的4个波形槽都设置默认波形
		for i := 0; i < 4; i++ {
			cmd.AWaveData[i] = defaultWave
			cmd.BWaveData[i] = defaultWave
		}
	}

	// 递增序列号
	c.sequence++
	// 序列号超过15则重置为1（序列号范围：1-15）
	if c.sequence > 15 {
		c.sequence = 1
	}

	// 返回构建好的B0指令
	return cmd
}

// ListPulses 列出可用波形
func (c *Controller) ListPulses() {
	pulses := c.pulseManager.ListPulses()
	fmt.Println("\n=== 可用波形 ===")
	for _, pulse := range pulses {
		marker := "  "
		if pulse.ID == c.channelState.CurrentPulse {
			marker = "* "
		}
		fmt.Printf("%s%s - %s\n", marker, pulse.ID, pulse.Name)
	}
	fmt.Println()
}

// AddStrength 增加通道强度
// TODO:NOTICE 增加强度 （通过蓝牙发送给设备指令）
func (c *Controller) AddStrength(channel string, value int) error {
	//channel是通道字符串 A还是B value是数值
	if !c.btAdapter.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	c.mu.Lock() // 加锁，确保线程安全
	defer c.mu.Unlock()

	// 获取当前强度
	var currentStrength int
	switch channel {
	case "A", "a":
		currentStrength = c.channelState.AStrength
	case "B", "b":
		currentStrength = c.channelState.BStrength
	default:
		return fmt.Errorf("无效的通道: %s", channel)
	}

	// 计算新强度
	newStrength := currentStrength + value
	if newStrength < 0 {
		newStrength = 0
	}

	// 验证强度值
	newStrength = int(protocol.ValidateStrength(newStrength)) // 验证强度值是否在有效范围内
	//这个有效范围在0到200之间

	// 检查上限
	var limit int
	switch channel {
	case "A", "a":
		limit = c.channelState.ALimit
		if newStrength > limit {
			newStrength = 20
			log.Printf("强度超过上限，调整为: %d", newStrength)
		}
		c.channelState.AStrength = newStrength
	case "B", "b":
		limit = c.channelState.BLimit
		if newStrength > limit {
			newStrength = 20
			log.Printf("强度超过上限，调整为: %d", newStrength)
		}
		c.channelState.BStrength = newStrength
	}

	// 构建并发送B0指令
	cmd := c.buildB0Command()
	switch channel {
	case "A", "a":
		cmd.AMode = protocol.StrengthModeAbsolute
		cmd.AStrength = byte(newStrength)
	case "B", "b":
		cmd.BMode = protocol.StrengthModeAbsolute
		cmd.BStrength = byte(newStrength)
	}

	log.Printf("增加%s通道强度%d，当前强度: %d", channel, value, newStrength)
	return c.sendCommand(cmd) //TODO:NOTICE 发送指令
}

// SubStrength 减少通道强度
func (c *Controller) SubStrength(channel string, value int) error {
	if !c.btAdapter.IsConnected() {
		return fmt.Errorf("设备未连接")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 获取当前强度
	var currentStrength int
	switch channel {
	case "A", "a":
		currentStrength = c.channelState.AStrength
	case "B", "b":
		currentStrength = c.channelState.BStrength
	default:
		return fmt.Errorf("无效的通道: %s", channel)
	}

	// 计算新强度
	newStrength := currentStrength - value
	if newStrength < 0 {
		newStrength = 0
	}

	// 验证强度值
	newStrength = int(protocol.ValidateStrength(newStrength))

	// 更新状态
	switch channel {
	case "A", "a":
		c.channelState.AStrength = newStrength
	case "B", "b":
		c.channelState.BStrength = newStrength
	}

	// 构建并发送B0指令
	cmd := c.buildB0Command()
	switch channel {
	case "A", "a":
		cmd.AMode = protocol.StrengthModeAbsolute
		cmd.AStrength = byte(newStrength)
	case "B", "b":
		cmd.BMode = protocol.StrengthModeAbsolute
		cmd.BStrength = byte(newStrength)
	}

	log.Printf("减少%s通道强度%d，当前强度: %d", channel, value, newStrength)
	return c.sendCommand(cmd)
}

// GetStatus 获取设备当前状态
func (c *Controller) GetStatus() *ChannelState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 返回通道状态的副本，避免外部直接修改
	return &ChannelState{
		AStrength:    c.channelState.AStrength,
		BStrength:    c.channelState.BStrength,
		ALimit:       c.channelState.ALimit,
		BLimit:       c.channelState.BLimit,
		CurrentPulse: c.channelState.CurrentPulse,
		BatteryLevel: c.channelState.BatteryLevel,
	}
}

// GetPulseList 获取可用波形列表
func (c *Controller) GetPulseList() []*pulse.PulseData {
	return c.pulseManager.ListPulses()
}

// IsConnected 获取设备连接状态
func (c *Controller) IsConnected() bool {
	return c.btAdapter.IsConnected()
}
