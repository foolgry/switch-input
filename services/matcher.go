package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Rule 输入法切换规则
type Rule struct {
	AppName    string `json:"app"`        // 应用程序名称或包名
	WindowName string `json:"window"`     // 窗口名称模式（可选）
	Input      string `json:"input"`      // 目标输入法ID
	Enabled    bool   `json:"enabled"`    // 是否启用
	Priority   int    `json:"priority"`   // 优先级（数字越小优先级越高）
}

// Config 配置文件结构
type Config struct {
	Rules         []Rule       `json:"rules"`         // 切换规则
	General       GeneralConfig `json:"general"`       // 通用配置
	LastModified  string       `json:"lastModified"`  // 最后修改时间
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	AutoStart       bool          `json:"autoStart"`       // 开机自启
	CheckInterval   int           `json:"checkInterval"`   // 检查间隔（毫秒）
	SwitchDelay     int           `json:"switchDelay"`     // 切换延迟（毫秒）
	EnableLogging   bool          `json:"enableLogging"`   // 启用日志
	LogLevel        string        `json:"logLevel"`        // 日志级别
	ShowNotifications bool       `json:"showNotifications"` // 显示通知
}

// MatcherService 规则匹配服务
type MatcherService struct {
	config     *Config
	configPath string
	ruleMap    map[string][]Rule // appName -> rules
	ruleMutex  sync.RWMutex
	onRuleMatch func(*Rule, *WindowInfo)
}

// NewMatcherService 创建新的规则匹配服务
func NewMatcherService(configPath string) *MatcherService {
	return &MatcherService{
		configPath: configPath,
		ruleMap:    make(map[string][]Rule),
	}
}

// LoadConfig 加载配置文件
func (ms *MatcherService) LoadConfig() error {
	ms.ruleMutex.Lock()
	defer ms.ruleMutex.Unlock()

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(ms.configPath); os.IsNotExist(err) {
		if err := ms.createDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create default config: %v", err)
		}
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(ms.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// 设置默认值
	if config.General.CheckInterval == 0 {
		config.General.CheckInterval = 500
	}
	if config.General.SwitchDelay == 0 {
		config.General.SwitchDelay = 100
	}
	if config.General.LogLevel == "" {
		config.General.LogLevel = "info"
	}

	ms.config = &config
	ms.buildRuleMap()

	return nil
}

// createDefaultConfig 创建默认配置文件
func (ms *MatcherService) createDefaultConfig() error {
	defaultConfig := &Config{
		Rules: []Rule{
			{
				AppName:  "com.apple.Safari",
				Input:    "com.tencent.inputmethod.wetype.pinyin",
				Enabled:  true,
				Priority: 1,
			},
			{
				AppName:  "com.google.Chrome",
				Input:    "com.tencent.inputmethod.wetype.pinyin",
				Enabled:  true,
				Priority: 1,
			},
			{
				AppName:  "com.apple.Terminal",
				Input:    "com.apple.keylayout.ABC",
				Enabled:  true,
				Priority: 1,
			},
			{
				AppName:  "com.microsoft.VSCode",
				Input:    "com.apple.keylayout.ABC",
				Enabled:  true,
				Priority: 1,
			},
		},
		General: GeneralConfig{
			AutoStart:        false,
			CheckInterval:    500,
			SwitchDelay:      100,
			EnableLogging:    true,
			LogLevel:         "info",
			ShowNotifications: true,
		},
	}

	return ms.SaveConfig(defaultConfig)
}

// SaveConfig 保存配置文件
func (ms *MatcherService) SaveConfig(config *Config) error {
	ms.ruleMutex.Lock()
	defer ms.ruleMutex.Unlock()

	// 确保目录存在
	dir := filepath.Dir(ms.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// 更新最后修改时间
	config.LastModified = fmt.Sprintf("%d", os.Getpid())

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// 写入文件
	if err := ioutil.WriteFile(ms.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	ms.config = config
	ms.buildRuleMap()

	return nil
}

// buildRuleMap 构建规则映射表
func (ms *MatcherService) buildRuleMap() {
	ms.ruleMap = make(map[string][]Rule)

	if ms.config == nil {
		return
	}

	for _, rule := range ms.config.Rules {
		if !rule.Enabled {
			continue
		}

		// 支持多个应用名称，用逗号分隔
		appNames := strings.Split(rule.AppName, ",")
		for _, appName := range appNames {
			appName = strings.TrimSpace(appName)
			if appName != "" {
				ms.ruleMap[appName] = append(ms.ruleMap[appName], rule)
			}
		}
	}

	// 对每个应用的规则按优先级排序
	for appName := range ms.ruleMap {
		rules := ms.ruleMap[appName]
		for i := 0; i < len(rules)-1; i++ {
			for j := i + 1; j < len(rules); j++ {
				if rules[i].Priority > rules[j].Priority {
					rules[i], rules[j] = rules[j], rules[i]
				}
			}
		}
	}
}

// MatchWindow 匹配窗口并返回对应的规则
func (ms *MatcherService) MatchWindow(window *WindowInfo) *Rule {
	ms.ruleMutex.RLock()
	defer ms.ruleMutex.RUnlock()

	if window == nil {
		return nil
	}

	// 首先尝试精确匹配应用名称
	if rules, exists := ms.ruleMap[window.AppName]; exists {
		for _, rule := range rules {
			if ms.windowNameMatches(window.WindowName, rule.WindowName) {
				return &rule
			}
		}
	}

	// 尝试模糊匹配
	for appName, rules := range ms.ruleMap {
		if ms.appNameMatches(window.AppName, appName) {
			for _, rule := range rules {
				if ms.windowNameMatches(window.WindowName, rule.WindowName) {
					return &rule
				}
			}
		}
	}

	return nil
}

// appNameMatches 应用名称匹配
func (ms *MatcherService) appNameMatches(currentAppName, ruleAppName string) bool {
	currentAppName = strings.ToLower(strings.TrimSpace(currentAppName))
	ruleAppName = strings.ToLower(strings.TrimSpace(ruleAppName))

	// 精确匹配
	if currentAppName == ruleAppName {
		return true
	}

	// 包含匹配
	if strings.Contains(currentAppName, ruleAppName) || strings.Contains(ruleAppName, currentAppName) {
		return true
	}

	// 包含关键词匹配
	ruleParts := strings.Split(ruleAppName, " ")
	for _, part := range ruleParts {
		if strings.TrimSpace(part) != "" && strings.Contains(currentAppName, part) {
			return true
		}
	}

	return false
}

// windowNameMatches 窗口名称匹配
func (ms *MatcherService) windowNameMatches(currentWindowName, ruleWindowName string) bool {
	// 如果规则中没有指定窗口名称，则匹配所有窗口
	if ruleWindowName == "" {
		return true
	}

	currentWindowName = strings.ToLower(strings.TrimSpace(currentWindowName))
	ruleWindowName = strings.ToLower(strings.TrimSpace(ruleWindowName))

	// 支持通配符匹配
	if strings.Contains(ruleWindowName, "*") {
		pattern := strings.ReplaceAll(ruleWindowName, "*", ".*")
		return strings.Contains(currentWindowName, strings.ReplaceAll(pattern, ".*", ""))
	}

	// 包含匹配
	if strings.Contains(currentWindowName, ruleWindowName) {
		return true
	}

	return false
}

// GetConfig 获取当前配置
func (ms *MatcherService) GetConfig() *Config {
	ms.ruleMutex.RLock()
	defer ms.ruleMutex.RUnlock()

	if ms.config == nil {
		return nil
	}

	// 返回配置的深拷贝
	configCopy := *ms.config
	return &configCopy
}

// AddRule 添加新规则
func (ms *MatcherService) AddRule(rule Rule) error {
	config := ms.GetConfig()
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	// 设置默认值
	if rule.Priority == 0 {
		rule.Priority = len(config.Rules) + 1
	}

	config.Rules = append(config.Rules, rule)
	return ms.SaveConfig(config)
}

// UpdateRule 更新规则
func (ms *MatcherService) UpdateRule(index int, rule Rule) error {
	config := ms.GetConfig()
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	if index < 0 || index >= len(config.Rules) {
		return fmt.Errorf("rule index out of range")
	}

	config.Rules[index] = rule
	return ms.SaveConfig(config)
}

// DeleteRule 删除规则
func (ms *MatcherService) DeleteRule(index int) error {
	config := ms.GetConfig()
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	if index < 0 || index >= len(config.Rules) {
		return fmt.Errorf("rule index out of range")
	}

	config.Rules = append(config.Rules[:index], config.Rules[index+1:]...)
	return ms.SaveConfig(config)
}

// SetRuleMatchCallback 设置规则匹配回调
func (ms *MatcherService) SetRuleMatchCallback(callback func(*Rule, *WindowInfo)) {
	ms.onRuleMatch = callback
}

// GetAppRules 获取特定应用的规则
func (ms *MatcherService) GetAppRules(appName string) []Rule {
	ms.ruleMutex.RLock()
	defer ms.ruleMutex.RUnlock()

	var appRules []Rule
	if rules, exists := ms.ruleMap[appName]; exists {
		for _, rule := range rules {
			appRules = append(appRules, rule)
		}
	}

	return appRules
}

// ReloadConfig 重新加载配置文件
func (ms *MatcherService) ReloadConfig() error {
	return ms.LoadConfig()
}