package services

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// InputMethod 输入法信息
type InputMethod struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// InputService 输入法管理服务
type InputService struct {
	currentInput string
}

// NewInputService 创建新的输入法管理服务
func NewInputService() *InputService {
	return &InputService{}
}

// GetCurrentInput 获取当前输入法
func (is *InputService) GetCurrentInput() (*InputMethod, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return is.getCurrentInputMac()
}

// getCurrentInputMac macOS下获取当前输入法
func (is *InputService) getCurrentInputMac() (*InputMethod, error) {
	cmd := exec.Command("/opt/homebrew/bin/im-select")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current input: %v", err)
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("no output from im-select")
	}

	inputID := strings.TrimSpace(string(output))
	return &InputMethod{
		ID:   inputID,
		Name: inputID,
	}, nil
}


// SwitchInput 切换到指定输入法
func (is *InputService) SwitchInput(inputID string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return is.switchInputMac(inputID)
}

// switchInputMac macOS下切换输入法
func (is *InputService) switchInputMac(inputID string) error {
	cmd := exec.Command("/opt/homebrew/bin/im-select", inputID)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to switch input: %v", err)
	}

	return nil
}


// GetAvailableInputs 获取可用的输入法列表
func (is *InputService) GetAvailableInputs() ([]*InputMethod, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return is.getAvailableInputsMac()
}

// getAvailableInputsMac macOS下获取可用输入法
func (is *InputService) getAvailableInputsMac() ([]*InputMethod, error) {
	inputs := []*InputMethod{}
	inputs = append(inputs, &InputMethod{
		ID:   "com.apple.keylayout.ABC",
		Name: "ABC",
	})
	inputs = append(inputs, &InputMethod{
		ID:   "com.tencent.inputmethod.wetype.pinyin",
		Name: "拼音",
	})

	return inputs, nil
}

