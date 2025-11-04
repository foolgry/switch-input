package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	AppName   string    `json:"appName,omitempty"`
	Input     string    `json:"input,omitempty"`
	Action    string    `json:"action,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// LoggerService 日志服务
type LoggerService struct {
	logFile      *os.File
	logPath      string
	maxLogSize   int64  // 最大日志文件大小（字节）
	maxLogFiles  int    // 保留的日志文件数量
	bufferSize   int    // 缓冲区大小
	logBuffer    []LogEntry
	bufferMutex  sync.RWMutex
	flushTicker  *time.Ticker
	stopChan     chan bool
	enableLogging bool
}

// NewLoggerService 创建新的日志服务
func NewLoggerService(logPath string) *LoggerService {
	return &LoggerService{
		logPath:      logPath,
		maxLogSize:   10 * 1024 * 1024, // 10MB
		maxLogFiles:  5,
		bufferSize:   100,
		logBuffer:    make([]LogEntry, 0, 100),
		stopChan:     make(chan bool),
		enableLogging: true,
	}
}

// Start 启动日志服务
func (ls *LoggerService) Start() error {
	// 确保日志目录存在
	dir := filepath.Dir(ls.logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// 打开日志文件
	var err error
	ls.logFile, err = os.OpenFile(ls.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// 启动定时刷新
	ls.flushTicker = time.NewTicker(5 * time.Second)
	go ls.flushRoutine()

	return nil
}

// Stop 停止日志服务
func (ls *LoggerService) Stop() {
	if ls.flushTicker != nil {
		ls.flushTicker.Stop()
	}

	close(ls.stopChan)
	ls.flush() // 确保所有日志都被写入

	if ls.logFile != nil {
		ls.logFile.Close()
	}
}

// SetLogging 设置是否启用日志
func (ls *LoggerService) SetLogging(enabled bool) {
	ls.enableLogging = enabled
}

// LogDebug 记录调试日志
func (ls *LoggerService) LogDebug(message string) {
	if !ls.enableLogging {
		return
	}
	ls.log(LogLevelDebug, message, "", "", "", "")
}

// LogInfo 记录信息日志
func (ls *LoggerService) LogInfo(message string) {
	if !ls.enableLogging {
		return
	}
	ls.log(LogLevelInfo, message, "", "", "", "")
}

// LogWarn 记录警告日志
func (ls *LoggerService) LogWarn(message string) {
	if !ls.enableLogging {
		return
	}
	ls.log(LogLevelWarn, message, "", "", "", "")
}

// LogError 记录错误日志
func (ls *LoggerService) LogError(message string) {
	if !ls.enableLogging {
		return
	}
	ls.log(LogLevelError, message, "", "", "", "")
}

// LogWindowChange 记录窗口变化日志
func (ls *LoggerService) LogWindowChange(appName, windowName string) {
	if !ls.enableLogging {
		return
	}
	ls.log(LogLevelInfo, fmt.Sprintf("窗口切换: %s (%s)", appName, windowName), appName, "", "window_change", "")
}

// LogInputSwitch 记录输入法切换日志
func (ls *LoggerService) LogInputSwitch(appName, inputId, action string, err error) {
	if !ls.enableLogging {
		return
	}

	message := fmt.Sprintf("输入法切换: %s -> %s", appName, inputId)
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
		message += fmt.Sprintf(" (失败: %s)", errorMsg)
	}

	ls.log(LogLevelInfo, message, appName, inputId, action, errorMsg)
}

// LogRuleMatch 记录规则匹配日志
func (ls *LoggerService) LogRuleMatch(appName, inputId string) {
	if !ls.enableLogging {
		return
	}
	ls.log(LogLevelInfo, fmt.Sprintf("规则匹配: %s -> %s", appName, inputId), appName, inputId, "rule_match", "")
}

// log 内部日志记录方法
func (ls *LoggerService) log(level LogLevel, message, appName, input, action, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     ls.levelToString(level),
		Message:   message,
		AppName:   appName,
		Input:     input,
		Action:    action,
		Error:     errorMsg,
	}

	ls.bufferMutex.Lock()
	ls.logBuffer = append(ls.logBuffer, entry)

	// 如果缓冲区满了，立即刷新
	if len(ls.logBuffer) >= ls.bufferSize {
		ls.flush()
	}
	ls.bufferMutex.Unlock()
}

// flush 刷新日志缓冲区到文件
func (ls *LoggerService) flush() {
	ls.bufferMutex.Lock()
	defer ls.bufferMutex.Unlock()

	if len(ls.logBuffer) == 0 {
		return
	}

	// 检查日志文件大小，如果超过限制则轮转
	if err := ls.rotateLogIfNeeded(); err != nil {
		fmt.Printf("Failed to rotate log: %v\n", err)
		return
	}

	// 写入日志条目
	for _, entry := range ls.logBuffer {
		jsonData, err := json.Marshal(entry)
		if err != nil {
			fmt.Printf("Failed to marshal log entry: %v\n", err)
			continue
		}

		if ls.logFile != nil {
			_, err = ls.logFile.WriteString(string(jsonData) + "\n")
			if err != nil {
				fmt.Printf("Failed to write log entry: %v\n", err)
			}
		}
	}

	// 清空缓冲区
	ls.logBuffer = ls.logBuffer[:0]

	// 确保数据写入磁盘
	if ls.logFile != nil {
		ls.logFile.Sync()
	}
}

// flushRoutine 定时刷新协程
func (ls *LoggerService) flushRoutine() {
	for {
		select {
		case <-ls.flushTicker.C:
			ls.flush()
		case <-ls.stopChan:
			return
		}
	}
}

// rotateLogIfNeeded 如果需要则轮转日志文件
func (ls *LoggerService) rotateLogIfNeeded() error {
	if ls.logFile == nil {
		return nil
	}

	// 获取当前文件大小
	stat, err := ls.logFile.Stat()
	if err != nil {
		return err
	}

	// 如果文件没有超过大小限制，不需要轮转
	if stat.Size() < ls.maxLogSize {
		return nil
	}

	// 关闭当前文件
	ls.logFile.Close()

	// 移动现有日志文件
	baseName := ls.logPath
	ext := filepath.Ext(baseName)
	baseNameWithoutExt := baseName[:len(baseName)-len(ext)]

	// 删除最旧的日志文件
	oldestLog := fmt.Sprintf("%s.%d%s", baseNameWithoutExt, ls.maxLogFiles, ext)
	os.Remove(oldestLog)

	// 移动现有日志文件
	for i := ls.maxLogFiles - 1; i >= 1; i-- {
		currentLog := fmt.Sprintf("%s.%d%s", baseNameWithoutExt, i, ext)
		newLog := fmt.Sprintf("%s.%d%s", baseNameWithoutExt, i+1, ext)
		if i == ls.maxLogFiles-1 {
			os.Remove(currentLog)
		} else {
			os.Rename(currentLog, newLog)
		}
	}

	// 移动当前日志文件
	backupLog := fmt.Sprintf("%s.1%s", baseNameWithoutExt, ext)
	os.Rename(ls.logPath, backupLog)

	// 创建新的日志文件
	ls.logFile, err = os.OpenFile(ls.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %v", err)
	}

	return nil
}

// GetRecentLogs 获取最近的日志条目
func (ls *LoggerService) GetRecentLogs(limit int) ([]LogEntry, error) {
	file, err := os.Open(ls.logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []LogEntry
	decoder := json.NewDecoder(file)

	// 简化实现：读取所有日志然后返回最后的limit条
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			continue
		}
		logs = append(logs, entry)
	}

	// 返回最近的日志
	if len(logs) > limit {
		return logs[len(logs)-limit:], nil
	}
	return logs, nil
}

// ClearLogs 清空日志文件
func (ls *LoggerService) ClearLogs() error {
	ls.bufferMutex.Lock()
	defer ls.bufferMutex.Unlock()

	// 清空缓冲区
	ls.logBuffer = ls.logBuffer[:0]

	// 清空文件
	if ls.logFile != nil {
		ls.logFile.Close()
	}

	err := os.Remove(ls.logPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// 重新创建文件
	ls.logFile, err = os.OpenFile(ls.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to recreate log file: %v", err)
	}

	return nil
}

// GetLogStats 获取日志统计信息
func (ls *LoggerService) GetLogStats() (map[string]int64, error) {
	stats := map[string]int64{
		"total_size": 0,
		"total_entries": 0,
		"debug_count": 0,
		"info_count": 0,
		"warn_count": 0,
		"error_count": 0,
	}

	file, err := os.Open(ls.logPath)
	if err != nil {
		return stats, err
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return stats, err
	}
	stats["total_size"] = fileInfo.Size()

	// 统计日志条目
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			continue
		}

		stats["total_entries"]++
		switch entry.Level {
		case "debug":
			stats["debug_count"]++
		case "info":
			stats["info_count"]++
		case "warn":
			stats["warn_count"]++
		case "error":
			stats["error_count"]++
		}
	}

	return stats, nil
}

// levelToString 将日志级别转换为字符串
func (ls *LoggerService) levelToString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}