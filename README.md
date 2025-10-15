<h1 align="center">gooze-starter</h1>

<p align="center"> 这是一个为了快速构建一个 Go 项目而生的 Go+Gin 项目脚手架</p>

<p align="center"> 🧠 旨在让 Go 服务的构建「更快、更清晰、更优雅」</p>

<p align="center"> 它是站在巨人的肩膀上而实现的，使用了大量优质的第三方包，介意请慎用</p>

可以先去 [gooze-starter 文档中心](https://github.com/soryetong/gooze-starter) 查看文档说明

## 特性
* **超低学习成本与模块化解耦**：基于流行的第三方库封装，内置大量服务与工具，把常用的MySQL、HTTP、Redis、Mongo等以模块化的方式按需加载，让你聚焦业务逻辑的实现。


* **多数据库驱动可自由切换**：支持 MySQL、Postgres、SQLite等多种数据库驱动。GORM 与 SQLX 可自由切换，既保留链式操作的便捷，也支持原生 SQL 查询，兼顾灵活性与性能。


* **完善的认证与权限系统**：内置 JWT 认证、Casbin 鉴权机制，轻松接入 RBAC/ABAC 权限模型。结合中间件控制请求访问，实现用户认证、接口权限等，适用于企业级后台系统。


* **CLI 工具与自动生成为开发提效**：内置 gooze 命令行工具，支持自动生成Model、路由、控制器等结构代码。通过内建的服务与工具，丰富的第三方包，提升你的开发效率，让你专注业务本身。

## 要求

- Go 1.24 或更高版本

## 使用

> 有两种方式可以快速上手

###  1. cli 生成

可以使用以下命令创建一个新的 Go 项目：

```bash
go install github.com/soryetong/gooze-cli@latest
```

然后，**进入你想存放的项目的目录中**，运行以下命令：

```bash
gooze-cli init
```

> 如果`go install` 成功，却提示找不到 `gooze-cli` 命令，那么需要先添加环境变量

运行该命令后，会提示你输入项目名、项目类型等，按照提示输入即可



运行完成后，会按照 Go 社区的项目最佳实践来生成一个优雅的 Go 项目结构，并生成相应的代码文件。


关于 `gooze-cli` 的更多信息，请查看 [GitHub-gooze-cli](https://github.com/soryetong/gooze-cli)

**如果你认为 `gooze-cli` 生成的目录结构你不满意，那你完全可以使用第二种自行生成**
<br>

### 2. 自行使用

它只需要简单的几步就可以快速创建一个项目

1. 初始化你的项目文件夹

```bash
go mod init your_project_name

cd your_project_name
```
2. 拉取 `gooze`

```bash
go get -u github.com/soryetong/gooze-starter
```

3. 创建 `main.go`

```go
package main

import (
	"github.com/soryetong/gooze-starter"
)

func main() {
	gooze.Run()
}
```

4. 更新依赖

```[sh] bash
go mod tidy
```

5. 其他的目录、文件，你都可以按照你的需求和爱好来创建

更多信息，请查看 [gooze-starter 文档中心](https://github.com/soryetong/gooze-starter)

## 贡献

如果您发现任何问题或有任何改进意见，请随时提出问题或提交拉取请求。非常欢迎您的贡献！

## 许可证

gooze-starter 是根据MIT许可证发布的。有关更多信息，请参见 [LICENSE](LICENSE) 文件。

## 备注:
-----

**感觉不错请star, 谢谢!**

**欢迎加入 gooze-starter QQ交流群:980872596，与大家共同成长**

提交bug及特性: https://github.com/soryetong/gooze-starter/issues

<br>