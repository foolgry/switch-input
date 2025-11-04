package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"
	"sync"
	"time"

	"switch-input/services"
)

// App struct
type App struct {
	ctx            context.Context
	windowService  *services.WindowService
	inputService   *services.InputService
	matcherService *services.MatcherService
	loggerService  *services.LoggerService
	isRunning      bool
	isRunningMutex sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	// 获取用户家目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "." // fallback to current directory
	}

	// 创建配置目录和日志目录的绝对路径
	configDir := filepath.Join(homeDir, ".switch-input")
	logDir := filepath.Join(configDir, "logs")

	// 确保目录存在
	os.MkdirAll(configDir, 0755)
	os.MkdirAll(logDir, 0755)

	configPath := filepath.Join(configDir, "config.json")
	logPath := filepath.Join(logDir, "app.log")

	return &App{
		windowService:  services.NewWindowService(),
		inputService:   services.NewInputService(),
		matcherService: services.NewMatcherService(configPath),
		loggerService:  services.NewLoggerService(logPath),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 启动日志服务
	if err := a.loggerService.Start(); err != nil {
		fmt.Printf("Failed to start logger: %v\n", err)
	}

	a.loggerService.LogInfo("应用程序启动")

	// 加载配置文件
	if err := a.matcherService.LoadConfig(); err != nil {
		errorMsg := fmt.Sprintf("加载配置文件失败: %v", err)
		a.loggerService.LogError(errorMsg)
		fmt.Printf("Failed to load config: %v\n", err)
		// 不显示对话框，只记录日志
		return
	}

	// 根据配置设置日志服务
	config := a.matcherService.GetConfig()
	if config != nil {
		a.loggerService.SetLogging(config.General.EnableLogging)
	}

	// 设置窗口变化回调
	a.windowService.SetWindowChangeCallback(a.onWindowChange)

	// 设置规则匹配回调
	a.matcherService.SetRuleMatchCallback(a.onRuleMatch)

	// 启动窗口监控
	go a.windowService.StartMonitoring()

	a.loggerService.LogInfo("输入法自动切换服务已启动")
	fmt.Println("输入法自动切换服务已启动")
}

// onWindowChange 窗口变化处理
func (a *App) onWindowChange(window *services.WindowInfo) {
	if window == nil {
		return
	}

	a.loggerService.LogWindowChange(window.AppName, window.WindowName)
	fmt.Printf("窗口切换: %s (%s)\n", window.AppName, window.WindowName)

	// 查找匹配的规则
	rule := a.matcherService.MatchWindow(window)
	if rule != nil {
		a.onRuleMatch(rule, window)
	}
}

// onRuleMatch 规则匹配处理
func (a *App) onRuleMatch(rule *services.Rule, window *services.WindowInfo) {
	config := a.matcherService.GetConfig()
	if config == nil {
		return
	}

	a.loggerService.LogRuleMatch(rule.AppName, rule.Input)

	// 延迟切换，避免频繁切换
	time.Sleep(time.Duration(config.General.SwitchDelay) * time.Millisecond)

	fmt.Printf("匹配规则: %s -> %s\n", rule.AppName, rule.Input)

	// 切换输入法
	if err := a.inputService.SwitchInput(rule.Input); err != nil {
		a.loggerService.LogInputSwitch(rule.AppName, rule.Input, "switch_failed", err)
		fmt.Printf("切换输入法失败: %v\n", err)
	} else {
		a.loggerService.LogInputSwitch(rule.AppName, rule.Input, "switch_success", nil)
		fmt.Printf("已切换到输入法: %s\n", rule.Input)
	}
}

// GetActiveWindow 获取当前活动窗口
func (a *App) GetActiveWindow() (*services.WindowInfo, error) {
	return a.windowService.GetActiveWindow()
}

// GetCurrentInput 获取当前输入法
func (a *App) GetCurrentInput() (*services.InputMethod, error) {
	return a.inputService.GetCurrentInput()
}

// GetAvailableInputs 获取可用输入法列表
func (a *App) GetAvailableInputs() ([]*services.InputMethod, error) {
	return a.inputService.GetAvailableInputs()
}

// SwitchInput 切换输入法
func (a *App) SwitchInput(inputID string) error {
	return a.inputService.SwitchInput(inputID)
}

// GetConfig 获取配置
func (a *App) GetConfig() (*services.Config, error) {
	config := a.matcherService.GetConfig()
	if config == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	return config, nil
}

// SaveConfig 保存配置
func (a *App) SaveConfig(config *services.Config) error {
	if err := a.matcherService.SaveConfig(config); err != nil {
		return err
	}

	fmt.Println("配置已保存")
	return nil
}

// AddRule 添加规则
func (a *App) AddRule(rule services.Rule) error {
	if err := a.matcherService.AddRule(rule); err != nil {
		return err
	}

	fmt.Printf("已添加规则: %s -> %s\n", rule.AppName, rule.Input)
	return nil
}

// UpdateRule 更新规则
func (a *App) UpdateRule(index int, rule services.Rule) error {
	if err := a.matcherService.UpdateRule(index, rule); err != nil {
		return err
	}

	fmt.Printf("已更新规则: %s -> %s\n", rule.AppName, rule.Input)
	return nil
}

// DeleteRule 删除规则
func (a *App) DeleteRule(index int) error {
	config := a.matcherService.GetConfig()
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	if index < 0 || index >= len(config.Rules) {
		return fmt.Errorf("规则索引超出范围")
	}

	rule := config.Rules[index]
	if err := a.matcherService.DeleteRule(index); err != nil {
		return err
	}

	fmt.Printf("已删除规则: %s -> %s\n", rule.AppName, rule.Input)
	return nil
}

// TestRule 测试规则
func (a *App) TestRule(rule services.Rule) (bool, *services.WindowInfo, error) {
	window, err := a.windowService.GetActiveWindow()
	if err != nil {
		return false, nil, err
	}

	matched := a.matcherService.MatchWindow(window)
	if matched != nil && matched.AppName == rule.AppName {
		return true, window, nil
	}

	return false, window, nil
}

// IsRunning 获取运行状态
func (a *App) IsRunning() bool {
	a.isRunningMutex.RLock()
	defer a.isRunningMutex.RUnlock()
	return a.isRunning
}

// SetRunning 设置运行状态
func (a *App) SetRunning(running bool) {
	a.isRunningMutex.Lock()
	defer a.isRunningMutex.Unlock()
	a.isRunning = running

	if running {
		go a.windowService.StartMonitoring()
		fmt.Println("输入法自动切换已启用")
	} else {
		a.windowService.StopMonitoring()
		fmt.Println("输入法自动切换已禁用")
	}
}

// onDomReady is called after front-end resources have been loaded
func (a *App) onDomReady(ctx context.Context) {
	// 这里可以添加DOM加载完成后的初始化逻辑
}

// onBeforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) onBeforeClose(ctx context.Context) (prevent bool) {
	// 允许应用正常退出
	return false
}

// onShutdown is called when the application is shutting down
func (a *App) onShutdown(ctx context.Context) {
	// 清理资源
	if a.loggerService != nil {
		a.loggerService.LogInfo("应用程序正在关闭")
		a.loggerService.Stop()
	}
	if a.windowService != nil {
		a.windowService.StopMonitoring()
	}
}


// QuitApp 退出应用程序
func (a *App) QuitApp() {
	// 在状态栏应用中，我们使用 systray.Quit() 来退出
	// 这个方法保留但不再使用
	fmt.Println("退出应用程序")
}

// GetRecentLogs 获取最近的日志条目
func (a *App) GetRecentLogs(limit int) ([]services.LogEntry, error) {
	if a.loggerService == nil {
		return nil, fmt.Errorf("logger service not initialized")
	}
	return a.loggerService.GetRecentLogs(limit)
}

// GetLogStats 获取日志统计信息
func (a *App) GetLogStats() (map[string]int64, error) {
	if a.loggerService == nil {
		return nil, fmt.Errorf("logger service not initialized")
	}
	return a.loggerService.GetLogStats()
}

// ClearLogs 清空日志文件
func (a *App) ClearLogs() error {
	if a.loggerService == nil {
		return fmt.Errorf("logger service not initialized")
	}

	err := a.loggerService.ClearLogs()
	if err != nil {
		return err
	}

	a.loggerService.LogInfo("日志文件已清空")
	return nil
}

// openConfigFile 使用系统默认编辑器打开配置文件
func (a *App) openConfigFile() {
	// 获取用户家目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("无法获取用户家目录: %v\n", err)
		return
	}

	configPath := filepath.Join(homeDir, ".switch-input", "config.json")

	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "darwin":
		// macOS: 使用 open 命令
		cmd = exec.Command("open", configPath)
	case "windows":
		// Windows: 使用 start 命令
		cmd = exec.Command("cmd", "/c", "start", "", configPath)
	default:
		// Linux: 使用 xdg-open 命令
		cmd = exec.Command("xdg-open", configPath)
	}

	err = cmd.Run()
	if err != nil {
		a.loggerService.LogError(fmt.Sprintf("无法打开配置文件: %v", err))
		fmt.Printf("无法打开配置文件: %v\n", err)
	}
}

// reloadConfig 重新加载配置文件
func (a *App) reloadConfig() {
	fmt.Println("正在重新加载配置文件...")

	// 重新加载配置
	if err := a.matcherService.ReloadConfig(); err != nil {
		errorMsg := fmt.Sprintf("重新加载配置文件失败: %v", err)
		a.loggerService.LogError(errorMsg)
		fmt.Printf("%s\n", errorMsg)
		return
	}

	// 根据新配置更新日志服务
	config := a.matcherService.GetConfig()
	if config != nil {
		a.loggerService.SetLogging(config.General.EnableLogging)
	}

	successMsg := "配置文件重新加载成功"
	a.loggerService.LogInfo(successMsg)
	fmt.Printf("%s\n", successMsg)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
