package vobj

// ModuleType 模块类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=ModuleType -linecomment
type ModuleType int

const (
	// ModuleTypeUnknown 未知
	ModuleTypeUnknown ModuleType = iota // 未知

	// ModuleTypeMenu 菜单模块
	ModuleTypeMenu // 菜单模块
)
