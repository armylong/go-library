package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type BaseCommand struct{}

var (
	rootCmd    = &cobra.Command{Use: "app"}
	defaultCmd *cobra.Command
)

// AddCliCommand 注册命令
func (t BaseCommand) AddCliCommand(c *cobra.Command) {
	rootCmd.AddCommand(c)
}

// SetDefaultCommand 设置默认命令
func (t BaseCommand) SetDefaultCommand(c *cobra.Command) {
	if defaultCmd != nil {
		fmt.Fprintln(os.Stderr, "Warning: default command already set")
		return
	}
	defaultCmd = c
}

// Go 入口函数
func Go(Register func(BaseCommand)) {
	Register(BaseCommand{})

	// 空命令 → 执行默认命令
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if defaultCmd == nil {
			return fmt.Errorf("no default command registered")
		}

		// 👇 最标准、最优雅的写法
		defaultCmd.SetArgs(args)
		return defaultCmd.ExecuteContext(cmd.Context())
	}

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
