package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler MCP请求处理器
type Handler struct {
	service *Service
}

// NewHandler 创建新的Handler实例
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// MCPMessage MCP协议消息
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError MCP错误
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool MCP工具定义
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// HandleRequest 处理MCP请求
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 设置SSE头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		// 处理SSE连接
		h.handleSSE(w, r)
		return
	}

	if r.Method == http.MethodPost {
		// 处理JSON-RPC请求
		h.handleJSONRPC(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleSSE 处理Server-Sent Events连接
func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 发送初始化消息
	initMsg := MCPMessage{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]interface{}{"protocolVersion": "2024-11-05"},
	}

	h.sendSSEMessage(w, initMsg)
	flusher.Flush()

	// 保持连接
	select {
	case <-r.Context().Done():
		return
	}
}

// handleJSONRPC 处理JSON-RPC请求
func (h *Handler) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	var msg MCPMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		h.sendJSONRPCError(w, nil, -32700, "Parse error")
		return
	}

	switch msg.Method {
	case "initialize":
		h.handleInitialize(w, msg)
	case "tools/list":
		h.handleToolsList(w, msg)
	case "tools/call":
		h.handleToolsCall(w, msg)
	default:
		h.sendJSONRPCError(w, msg.ID, -32601, "Method not found")
	}
}

// handleInitialize 处理初始化请求
func (h *Handler) handleInitialize(w http.ResponseWriter, msg MCPMessage) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "DG-LAB MCP Server",
			"version": "1.0.0",
		},
	}

	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}

	h.sendJSONRPCResponse(w, response)
}

// handleToolsList 处理工具列表请求
func (h *Handler) handleToolsList(w http.ResponseWriter, msg MCPMessage) {
	tools := []Tool{
		{
			Name:        "set_strength",
			Description: "设置通道强度",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{"type": "string", "enum": []string{"A", "B"}},
					"strength": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 200},
				},
				"required": []string{"channel", "strength"},
			},
		},
		{
			Name:        "set_limit",
			Description: "设置通道强度上限",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"channel": map[string]interface{}{"type": "string", "enum": []string{"A", "B"}},
					"limit": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 200},
				},
				"required": []string{"channel", "limit"},
			},
		},
		{
			Name:        "set_pulse",
			Description: "设置波形",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pulse_id": map[string]interface{}{"type": "string"},
				},
				"required": []string{"pulse_id"},
			},
		},
		{
			Name:        "get_status",
			Description: "获取设备状态",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "list_pulses",
			Description: "获取可用波形列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	result := map[string]interface{}{
		"tools": tools,
	}

	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result:  result,
	}

	h.sendJSONRPCResponse(w, response)
}

// handleToolsCall 处理工具调用请求
func (h *Handler) handleToolsCall(w http.ResponseWriter, msg MCPMessage) {
	params, ok := msg.Params.(map[string]interface{})
	if !ok {
		h.sendJSONRPCError(w, msg.ID, -32602, "Invalid params")
		return
	}

	toolName, ok := params["name"].(string)
	if !ok {
		h.sendJSONRPCError(w, msg.ID, -32602, "Missing tool name")
		return
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	var result interface{}
	var err error

	switch toolName {
	case "set_strength":
		result, err = h.callSetStrength(arguments)
	case "set_limit":
		result, err = h.callSetLimit(arguments)
	case "set_pulse":
		result, err = h.callSetPulse(arguments)
	case "get_status":
		result, err = h.callGetStatus(arguments)
	case "list_pulses":
		result, err = h.callListPulses(arguments)
	default:
		h.sendJSONRPCError(w, msg.ID, -32601, "Unknown tool")
		return
	}

	if err != nil {
		h.sendJSONRPCError(w, msg.ID, -32603, err.Error())
		return
	}

	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      msg.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("%v", result),
				},
			},
		},
	}

	h.sendJSONRPCResponse(w, response)
}

// 工具调用实现
func (h *Handler) callSetStrength(args map[string]interface{}) (interface{}, error) {
	channel, ok := args["channel"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid channel")
	}
	strength, ok := args["strength"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid strength")
	}

	err := h.service.SetStrength(channel, int(strength))
	if err != nil {
		return nil, err
	}

	return "强度设置成功", nil
}

func (h *Handler) callSetLimit(args map[string]interface{}) (interface{}, error) {
	channel, ok := args["channel"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid channel")
	}
	limit, ok := args["limit"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid limit")
	}

	err := h.service.SetLimit(channel, int(limit))
	if err != nil {
		return nil, err
	}

	return "上限设置成功", nil
}

func (h *Handler) callSetPulse(args map[string]interface{}) (interface{}, error) {
	pulseID, ok := args["pulse_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid pulse_id")
	}

	err := h.service.SetPulse(pulseID)
	if err != nil {
		return nil, err
	}

	return "波形设置成功", nil
}

func (h *Handler) callGetStatus(args map[string]interface{}) (interface{}, error) {
	status := h.service.GetStatus()
	return status, nil
}

func (h *Handler) callListPulses(args map[string]interface{}) (interface{}, error) {
	pulses := h.service.ListPulses()
	return pulses, nil
}

// 辅助函数
func (h *Handler) sendSSEMessage(w http.ResponseWriter, msg MCPMessage) {
	data, _ := json.Marshal(msg)
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func (h *Handler) sendJSONRPCResponse(w http.ResponseWriter, msg MCPMessage) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func (h *Handler) sendJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := MCPMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
