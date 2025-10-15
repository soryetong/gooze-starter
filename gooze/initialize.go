package gooze

import (
	"fmt"
	"time"

	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzcache"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var configFile string
var env string
var show bool

func init() {
	gzconsole.RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	gzconsole.RootCmd.PersistentFlags().StringVar(&env, "env", "", "env file")
	gzconsole.RootCmd.PersistentFlags().BoolVar(&show, "show", true, "Whether to display startup information")
	gzconsole.RootCmd.CompletionOptions.DisableDefaultCmd = true
	gzconsole.RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initialize(configFile, env, show)
	}
}

func initialize(configFile, env string, show bool) error {
	if configFile == "" {
		return fmt.Errorf("config file is required")
	}

	if show {
		gzconsole.Show(getCommands(), getGlobalFlags())
	}

	// 1. 初始化配置文件
	if err := LoadConfig(configFile, env, &Config); err != nil {
		gzconsole.Echo = initSugaredLogger("")
		return err
	}

	// 2. 初始化 Echo 输出
	gzconsole.Echo = initSugaredLogger(Config.App.Env)

	// 3. 初始化日志
	initILog()

	// 4. 初始化缓存模块
	Cache = gzcache.New(
		viper.GetInt("App.CacheCap"),
		viper.GetInt("App.CacheShard"),
		time.Duration(viper.GetInt("App.CacheClear")),
	)

	return nil
}

func initSugaredLogger(env string) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	if !gzutil.InArray(env, []string{"dev", "local", "debug", "test"}) {
		config.OutputPaths = []string{}
	}
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeCaller = nil
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.DisableStacktrace = true
	logger, _ := config.Build()

	return logger.Sugar()
}

func getCommands() []gzconsole.CommandInfo {
	var commands = []gzconsole.CommandInfo{}
	for _, command := range gzconsole.RootCmd.Commands() {
		if !command.Hidden {
			commands = append(commands, gzconsole.CommandInfo{
				Name: command.Name(),
				Desc: command.Long,
			})
		}
	}
	return commands
}

func getGlobalFlags() []gzconsole.CommandInfo {
	var globalFlags = []gzconsole.CommandInfo{}
	gzconsole.RootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		line := "--" + f.Name
		if f.Shorthand != "" {
			line += ", -" + f.Shorthand
		}
		desc := f.Usage
		if f.DefValue != "" {
			desc += fmt.Sprintf(" (default: %s)", f.DefValue)
		}

		globalFlags = append(globalFlags, gzconsole.CommandInfo{
			Name: line,
			Desc: desc,
		})
	})

	return globalFlags
}
