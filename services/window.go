package services

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// WindowInfo 窗口信息结构
type WindowInfo struct {
	AppName    string `json:"appName"`
	AppPath    string `json:"appPath"`
	WindowName string `json:"windowName"`
	PID        int    `json:"pid"`
}

// WindowService 窗口检测服务
type WindowService struct {
	lastWindow     *WindowInfo
	checkInterval  time.Duration
	onWindowChange func(*WindowInfo)
	stopChan       chan bool
}

// NewWindowService 创建新的窗口检测服务
func NewWindowService() *WindowService {
	return &WindowService{
		checkInterval: 500 * time.Millisecond, // 每500ms检查一次
		stopChan:      make(chan bool),
	}
}

// GetActiveWindow 获取当前活动窗口信息
func (ws *WindowService) GetActiveWindow() (*WindowInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return ws.getActiveWindowMac()
	case "windows":
		return ws.getActiveWindowWindows()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getActiveWindowMac macOS下获取活动窗口
func (ws *WindowService) getActiveWindowMac() (*WindowInfo, error) {
	// 使用AppleScript获取前台应用信息 - 更简单的版本
	script := `tell application "System Events" to get name of first application process whose frontmost is true`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window: %v", err)
	}

	appName := strings.TrimSpace(string(output))
	if appName == "" {
		return nil, fmt.Errorf("no active application found")
	}

	// 获取进程ID
	scriptPID := `tell application "System Events" to unix id of first application process whose frontmost is true`
	cmd = exec.Command("osascript", "-e", scriptPID)
	outputPID, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process ID: %v", err)
	}

	var pid int
	_, err = fmt.Sscanf(strings.TrimSpace(string(outputPID)), "%d", &pid)
	if err != nil {
		pid = 0
	}

	return &WindowInfo{
		AppName:    appName,
		AppPath:    "",
		WindowName: "",
		PID:        pid,
	}, nil
}

// getActiveWindowWindows Windows下获取活动窗口
func (ws *WindowService) getActiveWindowWindows() (*WindowInfo, error) {
	// 使用PowerShell获取活动窗口信息
	script := `
Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
public class WindowInfo {
    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool GetWindowText(IntPtr hWnd, System.Text.StringBuilder lpString, int nMaxCount);

    [DllImport("user32.dll")]
    public static extern IntPtr GetForegroundWindow();
}
"@

$hwnd = [WindowInfo]::GetForegroundWindow()
$processId = 0
[WindowInfo]::GetWindowThreadProcessId($hwnd, [ref]$processId)

$process = Get-Process -Id $processId -ErrorAction SilentlyContinue
if ($process) {
    $windowTitle = New-Object System.Text.StringBuilder 256
    [WindowInfo]::GetWindowText($hwnd, $windowTitle, $windowTitle.Capacity)
    Write-Output ($process.ProcessName + "||" + $process.Path + "||" + $windowTitle.ToString() + "||" + $process.Id)
}
`

	cmd := exec.Command("powershell", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, fmt.Errorf("no active window found")
	}

	parts := strings.Split(outputStr, "||")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid output format")
	}

	// 解析PID
	var pid int
	_, err = fmt.Sscanf(parts[3], "%d", &pid)
	if err != nil {
		pid = 0
	}

	return &WindowInfo{
		AppName:    parts[0],
		AppPath:    parts[1],
		WindowName: parts[2],
		PID:        pid,
	}, nil
}

// StartMonitoring 开始监控窗口变化
func (ws *WindowService) StartMonitoring() {
	ticker := time.NewTicker(ws.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if window, err := ws.GetActiveWindow(); err == nil {
				if ws.lastWindow == nil || ws.lastWindow.AppName != window.AppName {
					ws.lastWindow = window
					if ws.onWindowChange != nil {
						ws.onWindowChange(window)
					}
				}
			}
		case <-ws.stopChan:
			return
		}
	}
}

// StopMonitoring 停止监控
func (ws *WindowService) StopMonitoring() {
	close(ws.stopChan)
}

// SetWindowChangeCallback 设置窗口变化回调
func (ws *WindowService) SetWindowChangeCallback(callback func(*WindowInfo)) {
	ws.onWindowChange = callback
}

// SetCheckInterval 设置检查间隔
func (ws *WindowService) SetCheckInterval(interval time.Duration) {
	ws.checkInterval = interval
}