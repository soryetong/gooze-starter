package gzconsole

import (
	"sort"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type startupTask struct {
	Priority int
	Name     string
	Cmd      *cobra.Command
}

var (
	startupTasks []startupTask
	Echo         *zap.SugaredLogger
	RootCmd      = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunStartupTasks()
		},
	}
)

func Register(priority int, cmd *cobra.Command) {
	RootCmd.AddCommand(cmd)
	startupTasks = append(startupTasks, startupTask{
		Priority: priority,
		Name:     cmd.Name(),
		Cmd:      cmd,
	})
}

func RunStartupTasks() error {
	// 1. 根据优先级排序，数字越大越先执行
	sort.SliceStable(startupTasks, func(i, j int) bool {
		return startupTasks[i].Priority > startupTasks[j].Priority
	})

	// 2. 按顺序执行任务，任何一个任务失败，立即中止
	for _, task := range startupTasks {
		if err := task.Cmd.RunE(task.Cmd, []string{}); err != nil {
			return err
		}
	}

	return nil
}
