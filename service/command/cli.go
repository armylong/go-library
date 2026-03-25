package command

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

// 空结构体，保持和原来一样的用法
type BaseCommand struct{}

var (
	rootCmd       = &cobra.Command{Use: "app"}
	defaultAction func() // 存储默认动作（serve/Web）
)

// AddCliCommand 注册命令
// 识别 serve 为默认启动命令
func (t BaseCommand) AddCliCommand(c *cobra.Command) {
	// 关键逻辑：命令名是 serve → 设置为默认启动
	if c.Name() == "serve" || c.Name() == "" {
		defaultAction = func() {
			_ = c.RunE(nil, nil)
		}
	}

	rootCmd.AddCommand(c)
}

// Go 入口函数，和原来完全一样调用方式
func Go(Register func(BaseCommand)) {
	// 执行注册（RegisterWeb / RegisterCmd 等）
	Register(BaseCommand{})

	// ======================
	// 核心：空命令 → 执行默认动作（启动Web）
	// ======================
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if defaultAction != nil {
			defaultAction()
		}
	}

	// ======================
	// 启动 Cobra
	// ======================
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}
