package protocol

import (
	"encoding/hex"
	"fmt"
)

// StrengthMode 强度值解读方式
type StrengthMode byte

const (
	StrengthModeNoChange StrengthMode = 0b00 // 不改变
	StrengthModeIncrease StrengthMode = 0b01 // 相对增加
	StrengthModeDecrease StrengthMode = 0b10 // 相对减少
	StrengthModeAbsolute StrengthMode = 0b11 // 绝对设置
)

// WaveData 波形数据
type WaveData struct {
	Frequency byte // 频率 (10-240)
	Strength  byte // 强度 (0-100)
}

// B0Command DG-LAB V3协议B0指令
type B0Command struct {
	Sequence  byte         // 序列号 (0-15)
	AMode     StrengthMode // A通道强度解读方式
	BMode     StrengthMode // B通道强度解读方式
	AStrength byte         // A通道强度设定值
	BStrength byte         // B通道强度设定值
	AWaveData [4]WaveData  // A通道波形数据(4组)
	BWaveData [4]WaveData  // B通道波形数据(4组)
}

// ToBytes 将B0指令转换为字节数组
func (cmd *B0Command) ToBytes() []byte {
	data := make([]byte, 20)

	// 指令头
	//第1字节：固定为 0xB0（指令标识）
	data[0] = 0xB0

	// 序列号和强度解读方式
	modeValue := byte(cmd.AMode)<<2 | byte(cmd.BMode)
	data[1] = (cmd.Sequence << 4) | modeValue
	//低4位是强度解读方式：
	// - 位3-2：A通道模式（来自AMode）
	// - 位1-0：B通道模式（来自BMode） 注意这不是强度是模式

	// 通道强度设定值 一个字节 8位
	//最大256 而上限200是硬件设备的限制
	data[2] = cmd.AStrength
	data[3] = cmd.BStrength
	//- AStrength = 20 BStrength = 0

	// A通道波形数据
	// 每个字节 8位 前4位是频率 后4位是强度
	// 频率：10-240 强度：0-100
	// 0x0A 0x00 0x00 0x00 0x00 0x00 0x00 0x00
	// 0x0A 0x00 0x00 0x00 0x00 0x00 0x00 0x00
	// 0x0A 0x00 0x00 0x00 0x00 0x00 0x00 0x00
	// 0x0A 0x00 0x00 0x00 0x00 0x00 0x00 0x00
	for i, wave := range cmd.AWaveData {
		data[4+i] = wave.Frequency
		data[8+i] = wave.Strength
	}

	// B通道波形数据
	for i, wave := range cmd.BWaveData {
		data[12+i] = wave.Frequency
		data[16+i] = wave.Strength
	}

	return data
}

// WaveDataFromHex 从十六进制字符串创建波形数据
func WaveDataFromHex(hexStr string) ([4]WaveData, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return [4]WaveData{}, err
	}

	if len(data) != 8 {
		return [4]WaveData{}, fmt.Errorf("波形数据长度错误，期望8字节，实际%d字节", len(data))
	}

	var waves [4]WaveData
	for i := 0; i < 4; i++ {
		// 修正：前4字节是频率，后4字节是强度
		frequency := data[i]
		strength := data[i+4]

		// 验证频率范围 (10-240)
		if frequency < 10 {
			frequency = 10
		} else if frequency > 240 {
			frequency = 240
		}

		// 验证强度范围 (0-100)
		if strength > 100 {
			strength = 100
		}

		waves[i] = WaveData{
			Frequency: frequency,
			Strength:  strength,
		}
	}

	return waves, nil
}

// ConvertFrequency 将输入频率值(10-1000)转换为协议频率值(10-240)
func ConvertFrequency(input int) byte {
	switch {
	case input >= 10 && input <= 100:
		return byte(input)
	case input >= 101 && input <= 600:
		return byte((input-100)/5 + 100)
	case input >= 601 && input <= 1000:
		return byte((input-600)/10 + 200)
	default:
		return 10 // 默认最小值
	}
}

// ValidateStrength 验证强度值是否有效
func ValidateStrength(strength int) byte {
	if strength < 0 {
		return 0
	}
	if strength > 200 {
		return 200
	}
	return byte(strength)
}
