# DGLAB蓝牙控制器MCP (DG-LAB MCP Server)
基于 DG-LAB V3 协议的郊狼设备蓝牙控制系统，支持 MCP (Model Context Protocol) 接口。

## 项目概述
本项目是一个用 Go 语言开发的郊狼设备控制器，通过蓝牙连接郊狼设备，提供 HTTP API 和 MCP 协议接口，支持设置通道强度、波形切换等功能。

## 功能特性
- 🔗 蓝牙连接 : 自动扫描并连接郊狼设备
- ⚡ 强度控制 : 支持 A/B 双通道强度设置 (0-200)
- 🌊 波形管理 : 内置多种波形模式（呼吸、潮汐、连击等）
- 🔧 MCP 协议 : 支持 Model Context Protocol 接口
- 🌐 HTTP API : RESTful API 接口
- 📊 实时状态 : 设备连接状态、电量监控
- 🛡️ 安全限制 : 可配置强度上限保护
## 系统架构
```
├── main.go                 # 主程序入口
├── config.yaml            # 配置文件
├── pulses.yaml            # 波形配置
├── internal/
│   ├── bluetooth/         # 蓝牙通信模块
│   │   └── adapter.go
│   ├── config/           # 配置管理
│   │   └── config.go
│   ├── coyote/           # 设备控制器
│   │   └── controller.go
│   ├── mcp/              # MCP 协议实现
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── types.go
│   ├── protocol/         # DG-LAB 协议
│   │   └── dglab.go
│   └── pulse/            # 波形管理
│       └── manager.go
```
## 快速开始
### 环境要求
- Go 1.24.1+
- Windows/Linux/macOS
- 蓝牙适配器
- 郊狼设备 (脉冲主机 3.0)
### 安装依赖
```
go mod download
```
### 配置设备
编辑 config.yaml 文件：

```
bluetooth:
  scan_timeout: 30
  device_names:
    - "47L121000"  # 脉冲主机 3.0
    - "47L120100"  # 无线传感器

channels:
  a_channel:
    enabled: true
    max_strength: 100
    default_strength: 0
  b_channel:
    enabled: false
    max_strength: 100
    default_strength: 0
```
### 运行程序
```
go run main.go
```
程序启动后会：

1. 1.
   自动扫描并连接郊狼设备
2. 2.
   启动 HTTP 服务器 (端口 8080)
3. 3.
   提供 MCP 协议接口
## API 使用
### MCP 协议调用 设置通道强度
```
curl -X POST http://localhost:8080/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "set_strength",
      "arguments": {
        "channel": "A",
        "strength": 20
      }
    }
  }'
``` 获取设备状态
```
curl -X POST http://localhost:8080/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_status",
      "arguments": {}
    }
  }'
``` 设置波形
```
curl -X POST http://localhost:8080/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "set_pulse",
      "arguments": {
        "pulse_id": "7eae1e5f"
      }
    }
  }'
```
### 可用的 MCP 工具
工具名称 描述 参数 set_strength 设置通道强度 channel : "A"/"B", strength : 0-200 set_limit 设置强度上限 channel : "A"/"B", limit : 0-200 set_pulse 设置波形 pulse_id : 波形ID get_status 获取设备状态 无参数 list_pulses 列出可用波形 无参数

### MCP客户端食用方法 运行程序后 MCP SETTING增加
```
   "DG-LABMCP":{
      "url":"http://localhost:8080/api/mcp"
    }
```
## 波形配置
系统内置多种波形模式，在 pulses.yaml 中配置：

```
- id: "d6f83af0"
  name: "呼吸"
  pulse_data:
    - "0A0A0A0A14141414"
    - "0A0A0A0A28282828"
    # ... 更多波形数据

- id: "7eae1e5f"
  name: "潮汐"
  pulse_data:
    - "0D0D0D0D0F0F0F0F"
    # ... 更多波形数据
```
### 内置波形
- 呼吸 ( d6f83af0 ): 渐强渐弱的呼吸节奏
- 潮汐 ( 7eae1e5f ): 如潮汐般的起伏波形
- 连击 ( eea0e4ce ): 连续脉冲刺激
## 开发指南
### 项目结构说明
- bluetooth : 蓝牙通信抽象层，处理设备扫描和连接
- config : 配置文件管理，支持 YAML 格式
- coyote : 核心控制器，实现设备控制逻辑
- mcp : MCP 协议实现，提供标准化接口
- protocol : DG-LAB V3 协议实现，处理底层通信
- pulse : 波形管理器，加载和管理波形数据
### 添加新波形
1. 1.
   在 pulses.yaml 中添加新的波形配置
2. 2.
   波形数据格式：16进制字符串，每8字节表示一组波形
3. 3.
   前4字节为频率，后4字节为强度
### 扩展 MCP 工具
1. 1.
   在 internal/mcp/handler.go 中添加新的工具定义
2. 2.
   实现对应的处理函数
3. 3.
   在 handleToolsCall 中添加路由
## 安全注意事项
⚠️ 重要提醒 ：

- 请在使用前仔细阅读设备说明书
- 建议从低强度开始测试
- 设置合理的强度上限
- 避免长时间高强度使用
- 如有不适请立即停止使用
## 故障排除
### 常见问题
设备连接失败

- 检查蓝牙是否开启
- 确认设备在扫描范围内
- 检查设备名称配置是否正确
强度设置无效

- 检查设备是否已连接
- 确认强度值在有效范围内 (0-200)
- 检查是否超过设置的上限
API 调用失败

- 确认服务器正在运行
- 检查请求格式是否正确
- 查看服务器日志获取详细错误信息
## 技术规格
- 协议版本 : DG-LAB V3
- 通信方式 : 蓝牙 BLE
- 强度范围 : 0-200
- 频率范围 : 10-240 Hz
- 波形通道 : A/B 双通道
- MCP 版本 : 2024-11-05
## 许可证
本项目采用开源许可证，具体请查看 LICENSE 文件。

## 贡献
欢迎提交 Issue 和 Pull Request 来改进项目。

<img width="1920" height="1040" alt="Snipaste_2025-08-19_16-19-27" src="https://github.com/user-attachments/assets/8966cdfc-7156-4652-ae22-661dc61d4cda" />
<img width="1508" height="920" alt="Snipaste_2025-08-19_16-23-58" src="https://github.com/user-attachments/assets/92b5696f-607c-4f62-a110-1079f1ac10f3" />
<img width="1404" height="933" alt="Snipaste_2025-08-19_16-26-42" src="https://github.com/user-attachments/assets/37ba464f-4413-4aeb-9f26-57c11aa7431a" />


