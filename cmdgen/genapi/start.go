package genapi

import (
	"fmt"
	"strings"

	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	src        string
	output     string
	requestLog bool
	handlerV2  bool
)

func init() {
	CmdGen.PersistentFlags().StringVar(&src, "src", "", "Path to API description")
	CmdGen.PersistentFlags().StringVar(&output, "output", "", "Output path for generated code")
	CmdGen.PersistentFlags().BoolVar(&requestLog, "log", false, "Open request Log")
	CmdGen.PersistentFlags().BoolVar(&handlerV2, "handlerV2", false, "Use handlerV2 with gen and custom")
}

var CmdGen = &cobra.Command{
	Use:   "api",
	Short: "Generate API handler & route",
	RunE: func(cmd *cobra.Command, args []string) error {
		moduleName, err := gzutil.GetModuleName()
		if err != nil {
			return fmt.Errorf("获取当前项目Module名失败, 错误为: [%v] \n", err)
		}
		if src == "" {
			return fmt.Errorf("输入api文件所在的路径")
		}
		if output == "" {
			return fmt.Errorf("输入生成的代码存放路径")
		}

		viper.SetDefault("App.RouterPrefix", "api/v1")
		viper.SetDefault("App.Addr", ":18168")
		gaGen := newGen(&gaContext{
			packageName:        moduleName,
			src:                src,
			output:             output,
			handlerV2:          handlerV2,
			needRequestLog:     requestLog,
			routerPrefix:       strings.TrimLeft(viper.GetString("App.RouterPrefix"), "/"),
			addr:               ":" + strings.TrimLeft(viper.GetString("App.Addr"), ":"),
			dtoPackageName:     "dto",
			logicPackageName:   "logic",
			logicFuncName:      make(map[string]string),
			logicName:          make(map[string]string),
			handlerPackageName: "handler",
			handlerName:        make(map[string]string),
			routerPackageName:  "router",
			serverPackageName:  "bootstrap",
		})
		if err = gaGen.Do(); err != nil {
			return fmt.Errorf("自动生成代码失败, 错误为: [%v]", err)
		}

		return nil
	},
}
