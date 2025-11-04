package main

import (
	"context"
	"log"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

var globalApp *App // 全局应用实例

func main() {
	// Create an instance of the app structure
	globalApp = NewApp()

	// 创建上下文用于应用服务
	ctx := context.Background()

	// 启动应用服务
	go func() {
		globalApp.startup(ctx)
	}()

	// 启动状态栏
	systray.Run(onReady, onExit)
}

// onReady 状态栏准备就绪时调用
func onReady() {
	// 设置状态栏图标（使用内置图标，实际使用时可以替换为自定义图标）
	systray.SetIcon(icon.Data)
	systray.SetTitle("")
	systray.SetTooltip("输入法自动切换服务")

	// 打开配置文件菜单项
	mOpenConfig := systray.AddMenuItem("打开配置文件", "使用系统默认编辑器打开配置文件")

	// 分隔线
	systray.AddSeparator()

	// 退出菜单项
	mQuit := systray.AddMenuItem("退出", "退出输入法自动切换应用")

	// 菜单项点击事件处理
	go func() {
		for {
			select {
			case <-mOpenConfig.ClickedCh:
				globalApp.openConfigFile()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

// onExit 状态栏退出时调用
func onExit() {
	// 清理资源
	if globalApp != nil {
		globalApp.onShutdown(context.Background())
	}
	log.Println("输入法自动切换应用已退出")
}
