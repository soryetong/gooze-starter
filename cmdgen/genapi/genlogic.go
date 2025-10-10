package genapi

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

const logicHeaderTemplate = `
package {{.PackageName}}

import (
	"context"
	{{if .HasDto}} "{{.DtoPackagePath}}" {{end}}
)

type {{.LogicName}} struct {
}

func New{{.LogicName}}() *{{.LogicName}} {
	return &{{.LogicName}}{}
}
`

const logicContentTemplate = `

// @Summary {{ .Summary }}
func (self *{{.LogicName}}) {{.FuncName}}(ctx context.Context,{{if .PathParam}} {{.PathParam}} int64,{{end}}{{if .RequestType}} params *{{.DtoPackageName}}.{{.RequestType}}{{end}}) ({{if .ResponseType}} resp {{if not (hasPrefix .ResponseType "[]")}}*{{.DtoPackageName}}.{{end}}{{.ResponseType}},{{end}} err error) {
    // TODO implement

    return
}
`

func (self *generator) GenLogic() (err error) {
	filename := filepath.Join(self.output, self.moduleName, self.logicPackageName, self.fileName)
	// filename := filepath.Join(self.logicPackagePath, self.fileName)
	if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	for _, service := range self.services {
		if service.Name == "" {
			continue
		}

		if err = self.combineLogicWrite(service, filename); err != nil {
			return err
		}
	}

	return nil
}

func (self *generator) combineLogicWrite(service *serviceSpec, filename string) error {
	logicName := fmt.Sprintf("%sLogic", self.fileFinalName)
	self.logicName[strings.ToLower(service.Name)] = logicName
	hasDto := false
	for _, route := range service.Routes {
		if route.ResponseType != "" || route.RequestType != "" {
			hasDto = true
			break
		}
	}

	var fileContent []byte
	if content, err := os.ReadFile(filename); err == nil {
		fileContent = content
	}

	if len(fileContent) == 0 {
		if err := self.writeHeaderTemplate(filename, logicName, self.logicPackageName, hasDto); err != nil {
			return err
		}
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		fileContent = content
	}

	for _, route := range service.Routes {
		funcName := gzutil.UcFirst(service.Name) + gzutil.UcFirst(route.Name)
		self.logicFuncName[strings.ToLower(service.Name)+strings.ToLower(route.Name)] = funcName
		if methodExists(string(fileContent), logicName, funcName) {
			gzconsole.Echo.Info(fmt.Sprintf("logic: %s 中 %s 方法已存在，不进行重写", filename, funcName))
			continue
		}

		logicData := map[string]interface{}{
			"Summary":        route.Summary,
			"LogicName":      logicName,
			"FuncName":       funcName,
			"RequestType":    route.RequestType,
			"ResponseType":   route.ResponseType,
			"DtoPackageName": self.dtoPackageName,
			"PathParam":      route.RustFulKey,
		}

		contentStr, err := executeTemplate(logicContentTemplate, logicData)
		if err != nil {
			return err
		}

		if err := appendToFile(filename, "\n"+contentStr); err != nil {
			return err
		}

		fileContent, _ = os.ReadFile(filename)
	}

	self.formatFileWithGofmt(filename)
	return nil
}

// 执行模板渲染
func executeTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// 写入 Header
func (self *generator) writeHeaderTemplate(filename, logicName, packageName string, hasDto bool) error {
	data := map[string]interface{}{
		"LogicName":      logicName,
		"PackageName":    packageName,
		"HasDto":         hasDto,
		"DtoPackagePath": filepath.ToSlash(filepath.Join(self.packageName, self.dtoPackagePath)),
	}
	content, err := executeTemplate(logicHeaderTemplate, data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(content), 0644)
}

// 检测方法是否已存在（更精确）
func methodExists(fileContent, logicName, funcName string) bool {
	pattern := fmt.Sprintf(`func\s+\(.*\*\s*%s\s*\)\s+%s\(`, regexp.QuoteMeta(logicName), regexp.QuoteMeta(funcName))
	match, _ := regexp.MatchString(pattern, fileContent)
	return match
}

// 附加到文件
func appendToFile(filename, content string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
