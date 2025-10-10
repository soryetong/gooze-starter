package genapi

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

type generator struct {
	*gaContext
}

type gaContext struct {
	packageName       string // 主包名
	src               string
	output            string
	routerPrefix      string
	addr              string
	needRequestLog    bool
	moduleName        string // 模块名称, 指的是 管理后台、C端APP、C端Web 这种
	basePackagePath   string
	fileName          string
	groupName         string
	nowFilePrefixName string
	fileFinalName     string

	dtoPackageName string
	dtoPackagePath string
	dtoContents    []*dtoContentsSpec

	logicPackageName string
	logicPackagePath string
	logicFuncName    map[string]string
	logicName        map[string]string
	services         []*serviceSpec

	handlerPackageName string
	handlerPackagePath string
	handlerName        map[string]string

	routerPackageName string
	routerPackagePath string

	serverPackageName string
	serverPackagePath string
}

type dtoContentsSpec struct {
	Name   string
	Fields []*dtoFieldSpec
}

type dtoFieldSpec struct {
	Name    string
	Type    string
	Tag     string
	Comment string
}

type serviceSpec struct {
	Name    string
	Group   string
	Summary string
	Routes  []*routeSpec
}

type routeSpec struct {
	Method       string
	Path         string
	RustFulKey   string
	Name         string
	RequestType  string
	ResponseType string
	Summary      string
}

func newGen(ctx *gaContext) *generator {
	return &generator{
		ctx,
	}
}

func (self *generator) Do() (err error) {
	info, err := os.Stat(self.src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		err = filepath.WalkDir(self.src, func(file string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() == false {
				filename := filepath.Base(file)
				if strings.HasPrefix(filename, ".") {
					return nil
				}

				err = self.start(file)
				if err != nil {
					return err
				}
			}

			return nil
		})

		return err
	}

	return self.start(self.src)
}

func trimBeforeKeyword(path, keyword string) string {
	index := strings.Index(path, keyword)
	if index == -1 {
		return ""
	}
	return path[index+len(keyword)+1:] // +1 to skip the `/`
}

func (self *generator) start(filename string) (err error) {
	gzconsole.Echo.Debugf("开始API文件: %s 内容读取", filename)

	newFilePath := gzutil.Ternary(filepath.IsAbs(filename), trimBeforeKeyword(filename, self.packageName), filename)
	filePathArr := strings.Split(newFilePath, string(filepath.Separator))
	switch len(filePathArr) {
	case 2:
		self.moduleName = ""
	case 3:
		self.moduleName = filePathArr[1]
	default:
		gzconsole.Echo.Warn(filename, "文件路径不合法, 将不解析\n")
		return nil
	}

	newOutput := gzutil.Ternary(filepath.IsAbs(self.output), trimBeforeKeyword(self.output, self.packageName), self.output)
	self.basePackagePath = filepath.Join(newOutput, self.moduleName)
	nowFilePrefixName := filepath.Base(strings.TrimSuffix(filename, filepath.Ext(filename)))
	self.fileName = nowFilePrefixName + ".go"
	self.nowFilePrefixName = nowFilePrefixName
	parts := strings.FieldsFunc(nowFilePrefixName, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == ':' || r == '_' || r == '-'
	})
	self.groupName = strings.Join(parts, "/")
	if len(parts) > 1 {
		var builder strings.Builder
		for _, part := range parts {
			builder.WriteString(gzutil.UcFirst(part))
		}
		self.fileFinalName = builder.String()
	} else {
		self.fileFinalName = gzutil.UcFirst(self.nowFilePrefixName)
	}

	fileContentByte, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成API文件解析\n")
	fileContent := string(fileContentByte)

	// Step 1: Analysis Data Transfer Objects.
	self.dtoPackagePath = filepath.Join(self.basePackagePath, self.dtoPackageName)
	if err = self.MatchDto(fileContent); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成DTO内容解析\n")

	// Step 2: Analysis Service.
	if err = self.MatchRoutesService(fileContent); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成Service服务解析\n")

	// Step 3: Generate Data Transfer Objects.
	if err = self.GenDto(); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成DTO代码生成\n")

	// Step 4: Generate Logic.
	self.logicPackagePath = filepath.Join(self.basePackagePath, self.logicPackageName)
	if err = self.GenLogic(); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成Logic代码生成\n")

	// Step 5: Generate Handler.
	self.handlerPackagePath = filepath.Join(self.basePackagePath, self.handlerPackageName)
	if err = self.GenHandler(); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成Handler代码生成\n")

	// Step 6: Generate Router.
	self.routerPackagePath = filepath.Join(self.basePackagePath, self.routerPackageName)
	if err = self.GenRouter(); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成Router代码生成\n")

	// Step 7: Generate Server.
	self.serverPackagePath = filepath.Join(self.basePackagePath, self.serverPackageName)
	if err = self.GenServer(); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成Server代码生成\n")

	// Step 8: Generate Swagger.
	if err = self.GenSwagger(); err != nil {
		return err
	}
	gzconsole.Echo.Debug("✅ 已完成Swagger代码生成\n")
	gzconsole.Echo.Infof("ℹ️ 提示: 文件: %s 代码生成已完成\n", filename)

	return nil
}

// formatFileWithGofmt formats a file using gofmt.
func (self *generator) formatFileWithGofmt(filepath string) {
	originalContent, err := os.ReadFile(filepath)
	if err != nil {
		// fmt.Printf("Error reading file %s: %v\n", filepath, err)
		return
	}

	// Run gofmt on the original content.
	cmd := exec.Command("gofmt")
	cmd.Stdin = bytes.NewReader(originalContent)
	var formattedContent bytes.Buffer
	cmd.Stdout = &formattedContent
	if err = cmd.Run(); err != nil {
		// fmt.Printf("Error running gofmt on file %s: %v\n", filepath, err)
		return
	}

	// Split lines and trim trailing empty lines
	formattedLines := bytes.Split(formattedContent.Bytes(), []byte("\n"))
	trimmedFormattedLines := trimTrailingEmptyLines(formattedLines)

	// Reassemble the final content
	var finalContent bytes.Buffer
	for i, line := range trimmedFormattedLines {
		finalContent.Write(line)
		if i < len(trimmedFormattedLines)-1 || len(line) > 0 { // Add newline except for the last empty line
			finalContent.WriteString("\n")
		}
	}

	// Write the formatted content back to the file.
	if err = os.WriteFile(filepath, finalContent.Bytes(), 0644); err != nil {
		// fmt.Printf("Error writing formatted content to file %s: %v\n", filepath, err)
	}
}

// trimTrailingEmptyLines removes trailing empty lines from a slice of lines.
func trimTrailingEmptyLines(lines [][]byte) [][]byte {
	end := len(lines)
	for end > 0 && len(bytes.TrimSpace(lines[end-1])) == 0 {
		end--
	}
	return lines[:end]
}
