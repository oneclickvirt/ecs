package api

import (
	"github.com/oneclickvirt/ecs/internal/menu"
	"github.com/oneclickvirt/ecs/utils"
)

// GetMenuChoice 获取用户菜单选择
// language: 语言 ("zh" 或 "en")
// 返回: 用户选择的选项
func GetMenuChoice(language string) string {
	return menu.GetMenuChoice(language)
}

// PrintMenuOptions 打印菜单选项
// preCheck: 网络检查结果
// config: 配置对象
func PrintMenuOptions(preCheck utils.NetCheckResult, config *Config) {
	menu.PrintMenuOptions(preCheck, config)
}

// HandleMenuMode 处理菜单模式
// preCheck: 网络检查结果
// config: 配置对象
func HandleMenuMode(preCheck utils.NetCheckResult, config *Config) {
	menu.HandleMenuMode(preCheck, config)
}
