package bluetooth // 定义蓝牙适配器包

import (
	"fmt"     // 导入格式化输出包
	"log"     // 导入日志记录包
	"strings" // 导入字符串处理包
	"time"    // 导入时间处理包

	"tinygo.org/x/bluetooth" // 导入TinyGo蓝牙库
)

// BluetoothAdapter 真实的蓝牙适配器结构体
type BluetoothAdapter struct {
	adapter        *bluetooth.Adapter             // 蓝牙适配器实例指针
	device         *bluetooth.Device              // 连接的蓝牙设备指针（改为指针类型）
	connected      bool                           // 连接状态标志
	serviceUUID    bluetooth.UUID                 // DG-LAB服务UUID
	charUUID       bluetooth.UUID                 // DG-LAB特征UUID
	characteristic bluetooth.DeviceCharacteristic // 设备特征实例
}

// NewBluetoothAdapter 创建新的蓝牙适配器实例
func NewBluetoothAdapter() *BluetoothAdapter {
	// DG-LAB V3协议的正确UUID（根据官方文档）
	serviceUUID, _ := bluetooth.ParseUUID("0000180c-0000-1000-8000-00805f9b34fb") // 解析DG-LAB主服务UUID (0x180C)
	charUUID, _ := bluetooth.ParseUUID("0000150a-0000-1000-8000-00805f9b34fb")    // 解析写入特征UUID (0x150A)

	return &BluetoothAdapter{ // 返回新的蓝牙适配器实例
		adapter:     bluetooth.DefaultAdapter, // 使用默认蓝牙适配器
		serviceUUID: serviceUUID,              // 设置服务UUID
		charUUID:    charUUID,                 // 设置特征UUID
	}
}

// Enable 启用蓝牙适配器
func (ba *BluetoothAdapter) Enable() error {
	log.Println("启用蓝牙适配器...")  // 记录启用开始日志
	err := ba.adapter.Enable() // 调用适配器启用方法 这是bluetooth.Adapter 蓝牙库里面的启用方法
	if err != nil {            // 检查是否有错误
		return fmt.Errorf("启用蓝牙失败: %w", err) // 返回格式化错误信息
	}
	log.Println("蓝牙适配器已启用") // 记录启用成功日志
	return nil              // 返回无错误
}

// TODO:NOTICE 蓝牙配对
// ScanAndConnect 扫描并连接到郊狼设备
func (ba *BluetoothAdapter) ScanAndConnect(timeout time.Duration, deviceNames []string) error {
	log.Println("开始扫描蓝牙设备...") // 记录扫描开始日志

	var targetDevice bluetooth.ScanResult // 声明目标设备变量  //蓝牙库里面的扫描对象
	found := false                        // 初始化找到标志为false

	// 扫描设备
	err := ba.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		deviceName := result.LocalName()
		//TODO:NOTICE 配对 脉冲主机     deviceName=  []string len: 2, cap: 2, ["47L121000","47L120100"]
		//     蓝牙名称 脉冲主机 3.0 : 47L121000   无线传感器 : 47L120100
		//   https://github.com/DG-LAB-OPENSOURCE/DG-LAB-OPENSOURCE/blob/main/coyote/v3/README_V3.md
		//  获取设备本地名称
		//deviceName 设备名 result.Address.String() MAC地址
		// log.Printf("发现设备: %s (%s) RSSI: %d", deviceName, result.Address.String(), result.RSSI) // 记录发现的设备信息

		// 检查是否是目标设备
		//TODO:NOTICE 寻找deviceName 其设备名和 deviceNames[index]的要一样 （deviceNames[index]是目标设备 deviceName是找到的设备）

		for _, name := range deviceNames { // 遍历目标设备名称列表
			if strings.Contains(strings.ToLower(deviceName), strings.ToLower(name)) { // 不区分大小写比较设备名称
				log.Printf("找到目标设备: %s", deviceName) // 记录找到目标设备日志
				targetDevice = result                // 保存目标设备扫描结果
				found = true                         // 设置找到标志为true
				ba.adapter.StopScan()                // 停止扫描
				return                               // 退出回调函数
			}
		}
	})

	if err != nil { // 检查扫描是否有错误
		return fmt.Errorf("扫描失败: %w", err) // 返回格式化错误信息
	}

	if !found { // 检查是否找到目标设备
		return fmt.Errorf("未找到目标设备，请确保设备已开启并处于可发现状态") // 返回未找到设备错误
	}

	// 连接到设备
	log.Printf("正在连接到设备: %s", targetDevice.Address.String())
	//TODO:NOTICE 连接设备要传递蓝牙的MAC地址
	// 记录连接开始日志
	//在这里传递一个空的 ConnectionParams{} 结构体意味着使用默认的连接参数进行连接，即不指定任何特殊的连接超时、MTU（最大传输单元）或其他高级连接选项。
	// 这对于大多数标准的蓝牙连接场景来说已经足够了。
	device, err := ba.adapter.Connect(targetDevice.Address, bluetooth.ConnectionParams{}) // 连接到目标设备
	if err != nil {                                                                       // 检查连接是否有错误
		return fmt.Errorf("连接设备失败: %w", err) // 返回格式化错误信息
	}

	ba.device = device    // 保存设备实例
	log.Println("设备连接成功") // 记录连接成功日志

	//TODO:NOTICE 发现DG-LAB主服务 (0x180C)
	//蓝牙设备通过"服务"（Service）来组织和提供不同的功能
	//每个服务都有一个唯一的UUID来标识，比如代码中的DG-LAB主服务UUID是 0x180C
	log.Println("正在发现服务...")                                                   // 记录服务发现开始日志
	services, err := device.DiscoverServices([]bluetooth.UUID{ba.serviceUUID}) // 发现指定的DG-LAB服务
	if err != nil {                                                            // 检查服务发现是否有错误
		return fmt.Errorf("发现服务失败: %w", err) // 返回格式化错误信息
	}

	if len(services) == 0 { // 检查是否找到服务
		return fmt.Errorf("未找到DG-LAB服务 (0x180C)") // 返回未找到服务错误
	}

	service := services[0]                                // 获取第一个（也是唯一的）服务
	log.Printf("找到DG-LAB服务: %s", service.UUID().String()) // 记录找到服务日志

	// 发现特征
	//TODO:NOTICE 在蓝牙通信中，特征（Characteristic）是服务（Service）的组成部分
	//  发现服务后，还需要知道这个服务提供哪些具体的数据交互点
	//程序需要找到正确的特征才能：- 发送命令（通过写入特征 0x150A） - 接收设备状态（通过通知特征 0x150B）

	log.Println("正在发现特征...")                                     // 记录特征发现开始日志
	characteristics, err := service.DiscoverCharacteristics(nil) // 发现所有特征（传入nil表示发现所有）
	if err != nil {                                              // 检查特征发现是否有错误
		return fmt.Errorf("发现特征失败: %w", err) // 返回格式化错误信息
	}

	log.Printf("发现了 %d 个特征", len(characteristics)) // 记录发现的特征数量

	//查找特征
	// 查找写入特征 (0x150A)
	var writeChar bluetooth.DeviceCharacteristic  // 声明写入特征变量
	var notifyChar bluetooth.DeviceCharacteristic // 声明通知特征变量
	writeCharFound := false                       // 初始化写入特征找到标志
	notifyCharFound := false                      // 初始化通知特征找到标志

	for _, char := range characteristics { // 遍历所有特征
		//在这个主机里面特征只有两个 写/通知
		//遍历 一下 如果不符合写入肯定就是通知 如果符合写入就是就写入
		charUUIDStr := char.UUID().String()   // 获取特征UUID字符串
		log.Printf("特征UUID: %s", charUUIDStr) // 记录特征UUID

		// 写入特征 (0x150A)
		if charUUIDStr == "0000150a-0000-1000-8000-00805f9b34fb" { // 检查是否是写入特征
			writeChar = char                      // 保存写入特征
			writeCharFound = true                 // 设置找到标志
			log.Printf("找到写入特征: %s", charUUIDStr) // 记录找到写入特征日志
		}

		// 通知特征 (0x150B)
		if charUUIDStr == "0000150b-0000-1000-8000-00805f9b34fb" { // 检查是否是通知特征
			notifyChar = char                     // 保存通知特征
			notifyCharFound = true                // 设置找到标志
			log.Printf("找到通知特征: %s", charUUIDStr) // 记录找到通知特征日志
		}
	}

	if !writeCharFound { // 检查是否找到写入特征
		return fmt.Errorf("未找到写入特征 (0x150A)") // 返回未找到写入特征错误
	}

	ba.characteristic = writeChar // 保存写入特征实例

	// 如果找到通知特征，启用通知
	if notifyCharFound { // 检查是否找到通知特征
		err = notifyChar.EnableNotifications(func(buf []byte) { // 启用通知并设置回调函数
			log.Printf("收到设备通知: %x", buf) // 记录收到的通知数据（十六进制格式）
		})
		if err != nil { // 检查启用通知是否有错误
			log.Printf("启用通知失败: %v", err) // 记录启用通知失败日志
		} else {
			log.Println("已启用设备通知") // 记录启用通知成功日志
		}
	}

	ba.connected = true          // 设置连接状态为true
	log.Println("DG-LAB设备连接完成！") // 记录连接完成日志
	return nil                   // 返回无错误
}

// WriteCharacteristic 写入特征值
// TODO:NOTICE 像设备写入指令
func (ba *BluetoothAdapter) WriteCharacteristic(data []byte) error {
	if !ba.connected { // 检查设备是否已连接
		return fmt.Errorf("设备未连接") // 返回设备未连接错误
	}

	log.Printf("发送数据: %x", data)                           // 记录发送的数据（十六进制格式）
	_, err := ba.characteristic.WriteWithoutResponse(data) // TODO:NOTICE 开电
	//向特征写入数据（无响应模式）
	if err != nil { // 检查写入是否有错误
		return fmt.Errorf("写入特征失败: %w", err) // 返回格式化错误信息
	}
	return nil // 返回无错误
}

// IsConnected 检查连接状态
func (ba *BluetoothAdapter) IsConnected() bool {
	return ba.connected // 返回当前连接状态
}

// Disconnect 断开连接
func (ba *BluetoothAdapter) Disconnect() error {
	if ba.device != nil && ba.connected { // 检查设备是否存在且已连接
		err := ba.device.Disconnect() // 断开设备连接
		ba.connected = false          // 设置连接状态为false
		ba.device = nil               // 清空设备实例
		return err                    // 返回断开连接的结果
	}
	return nil // 如果设备未连接，返回无错误
}
