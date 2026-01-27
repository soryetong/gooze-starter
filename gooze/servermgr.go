package gooze

import (
	"fmt"

	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var serviceMgrCmd = &cobra.Command{
	Use:    "Start",
	Short:  "Web 项目的服务启动",
	Long:   `通过注册你指定的路由启动一个 HTTP 服务`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		defer closeServiceMgr()
		if len(serviceList) <= 0 {
			return fmt.Errorf("请务必通过实现接口 `gooze.IService` 注册你要启动的服务")
		}

		var eg errgroup.Group
		for _, service := range serviceList {
			eg.Go(func() error {
				if err := service.OnStart(); err != nil {
					return fmt.Errorf("服务 %s: %v", gzutil.GetCallerName(service), err)
				}

				return nil
			})
		}

		// 等待所有任务完成
		if err := eg.Wait(); err != nil {
			return err
		}
		return nil
	},
}

func closeServiceMgr() {
	_ = gzconsole.Echo.Sync()
	_ = Log.Sync()
	if rotationSchedulerProcess != nil {
		rotationSchedulerProcess.Stop()
	}
}

var serviceList []IService

func RegisterService(service ...IService) {
	serviceList = append(serviceList, service...)
}

type IService interface {
	OnStart() error
}

type IServer struct {
	IService
}

func (self *IServer) OnStart() error {
	return nil
}
