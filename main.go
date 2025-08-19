package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mygodblab/internal/config"
	"mygodblab/internal/coyote"
	"mygodblab/internal/mcp"
)

func main() {
	fmt.Println("郊狼蓝牙控制器 v1.0.0")
	fmt.Println("基于DG-LAB V3协议")

	// 加载配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("加载配置失败，使用默认配置: %v", err)
		cfg = config.DefaultConfig()
	}

	// 创建郊狼控制器
	controller, err := coyote.NewController(cfg)
	if err != nil {
		log.Fatalf("创建控制器失败: %v", err)
	}
	defer controller.Close()

	// 启动蓝牙连接（在后台进行，不阻塞服务器启动）
	go func() {
		fmt.Println("正在扫描郊狼设备...")
		err = controller.ScanAndConnect(30 * time.Second)
		if err != nil {
			log.Printf("连接设备失败: %v", err)
			fmt.Println("设备未连接，但HTTP服务器仍可使用")
		} else {
			fmt.Println("设备连接成功！")
		}
	}()

	// 创建MCP服务
	service := mcp.NewService(controller)
	handler := mcp.NewHandler(service)

	// 设置HTTP路由
	http.HandleFunc("/api/mcp", handler.HandleRequest)

	// 启动HTTP服务器
	serverAddr := ":8080"
	fmt.Printf("MCP服务器启动在 http://localhost%s\n", serverAddr)
	fmt.Println("API端点: http://localhost:8080/api/mcp")
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

func runInteractiveMode(controller *coyote.Controller) {
	fmt.Println("\n可用命令:")
	fmt.Println("  set-strength <channel> <value>  - 设置通道强度 (0-200)")
	fmt.Println("  add-strength <channel> <value>  - 增加通道强度")
	fmt.Println("  sub-strength <channel> <value>  - 减少通道强度")
	fmt.Println("  set-limit <channel> <value>     - 设置通道强度上限")
	fmt.Println("  set-pulse <pulse_id>            - 更换波形")
	fmt.Println("  list-pulses                     - 列出可用波形")
	fmt.Println("  status                          - 显示当前状态")
	fmt.Println("  help                            - 显示帮助信息")
	fmt.Println("  quit                            - 退出程序")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin) // 从标准输入创建扫描器  开始输入

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := parts[0]

		switch cmd {
		case "quit", "exit":
			fmt.Println("再见！")
			return
		case "status":
			controller.PrintStatus()
		case "list-pulses":
			controller.ListPulses()
		case "help":
			showHelp()
		case "set-strength":
			handleSetStrength(controller, parts)
		case "add-strength":
			handleAddStrength(controller, parts)
		case "sub-strength":
			handleSubStrength(controller, parts)
		case "set-limit":
			handleSetLimit(controller, parts)
		case "set-pulse":
			handleSetPulse(controller, parts)
		default:
			fmt.Printf("未知命令: %s，输入 'help' 查看帮助\n", cmd)
		}
	}
}

func handleSetStrength(controller *coyote.Controller, parts []string) {
	if len(parts) != 3 {
		fmt.Println("用法: set-strength <channel> <value>")
		fmt.Println("示例: set-strength A 50")
		return
	}

	channel := parts[1]
	value, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Printf("无效的强度值: %s\n", parts[2])
		return
	}

	// 创建定时器，每100ms发送一次指令
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	fmt.Printf("开始持续输出到%s通道，强度为%d（按Ctrl+C停止）\n", channel, value)

	// 持续发送指令
	for range ticker.C {
		err = controller.SetStrength(channel, value)
		if err != nil {
			fmt.Printf("设置强度失败: %v\n", err)
			return
		}
	}
}

func handleAddStrength(controller *coyote.Controller, parts []string) {
	if len(parts) != 3 {
		fmt.Println("用法: add-strength <channel> <value>")
		fmt.Println("示例: add-strength A 10")
		return
	}

	channel := parts[1]
	value, err := strconv.Atoi(parts[2]) //Atoi 字符串转换为整数
	if err != nil {
		fmt.Printf("无效的强度值: %s\n", parts[2])
		return
	}

	err = controller.AddStrength(channel, value) //TODO:NOTICE 增加强度
	if err != nil {
		fmt.Printf("增加强度失败: %v\n", err)
	}
}

func handleSubStrength(controller *coyote.Controller, parts []string) {
	if len(parts) != 3 {
		fmt.Println("用法: sub-strength <channel> <value>")
		fmt.Println("示例: sub-strength A 5")
		return
	}

	channel := parts[1]
	value, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Printf("无效的强度值: %s\n", parts[2])
		return
	}

	err = controller.SubStrength(channel, value)
	if err != nil {
		fmt.Printf("减少强度失败: %v\n", err)
	}
}

func handleSetLimit(controller *coyote.Controller, parts []string) {
	if len(parts) != 3 {
		fmt.Println("用法: set-limit <channel> <value>")
		fmt.Println("示例: set-limit A 80")
		return
	}

	channel := parts[1]
	value, err := strconv.Atoi(parts[2])
	if err != nil {
		fmt.Printf("无效的上限值: %s\n", parts[2])
		return
	}

	err = controller.SetLimit(channel, value)
	if err != nil {
		fmt.Printf("设置上限失败: %v\n", err)
	}
}

func handleSetPulse(controller *coyote.Controller, parts []string) {
	if len(parts) != 2 {
		fmt.Println("用法: set-pulse <pulse_id>")
		fmt.Println("示例: set-pulse 7eae1e5f")
		fmt.Println("使用 'list-pulses' 查看可用波形")
		return
	}

	pulseID := parts[1]
	err := controller.SetPulse(pulseID)
	if err != nil {
		fmt.Printf("设置波形失败: %v\n", err)
	}
}

func showHelp() {
	fmt.Println("\n=== 命令帮助 ===")
	fmt.Println("set-strength <channel> <value>  - 设置通道强度")
	fmt.Println("  channel: A 或 B")
	fmt.Println("  value: 0-200")
	fmt.Println("  示例: set-strength A 50")
	fmt.Println()
	fmt.Println("add-strength <channel> <value>  - 增加通道强度")
	fmt.Println("  示例: add-strength A 10")
	fmt.Println()
	fmt.Println("sub-strength <channel> <value>  - 减少通道强度")
	fmt.Println("  示例: sub-strength A 5")
	fmt.Println()
	fmt.Println("set-limit <channel> <value>     - 设置通道强度上限")
	fmt.Println("  示例: set-limit A 80")
	fmt.Println()
	fmt.Println("set-pulse <pulse_id>            - 更换波形")
	fmt.Println("  示例: set-pulse 7eae1e5f")
	fmt.Println("  使用 'list-pulses' 查看可用波形ID")
	fmt.Println()
	fmt.Println("list-pulses                     - 列出所有可用波形")
	fmt.Println("status                          - 显示当前设备状态")
	fmt.Println("help                            - 显示此帮助信息")
	fmt.Println("quit                            - 退出程序")
	fmt.Println()
}
