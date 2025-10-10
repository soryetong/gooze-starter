package genapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

const handlerHeaderTemplate = `
package {{.PackageName}}

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gooze"
	{{- if .HasRestFul }}
	"github.com/spf13/cast"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
	{{- end }}
	{{- if .HasDto }}
	"github.com/soryetong/gooze-starter/pkg/gzerror"
	"{{ .DtoPackagePath }}"
	{{- end }}
	"{{ .LogicPackagePath }}"
)

var {{.LogicInstanceName}} = {{.LogicPackageName}}.New{{.LogicStructName}}()
`

const handlerContentTemplate = `

// @Summary {{ .Summary }}
// @Description {{ .Summary }}
// @Accept json
// @Produce json{{if .PathParam}}
// @Param id query int64 true "id"{{end}}
{{- if .RequestType }}
// @Param {{ if eq .Method "get" }}query{{ else }}body{{ end }} {{ .DtoPackageName }}.{{ .RequestType }}
{{- end }}
// @Success 200 {{if .ResponseType}}{object} {{.DtoPackageName}}.{{.ResponseType}} {{ else }}string success {{ end }}
// @Failure 200 {object} gooze.Response 根据Code表示不同类型的错误
// @Router {{ .Path }} [{{ .Method}}]
func {{ .HandlerName }}(ctx *gin.Context) {
{{if .PathParam}} id := cast.ToInt64(ctx.Param("{{.PathParam}}"))
	if !gzutil.IsValidNumber(id) {
		gooze.FailWithMessage(ctx, "参数错误")
		return
	}
{{end}}{{if .RequestType}}	var req {{.DtoPackageName}}.{{.RequestType}}
	if err := ctx.ShouldBind(&req); err != nil {
		gooze.FailWithMessage(ctx, gzerror.Trans(err))
		return
	}

	{{if .ResponseType}}resp, err := {{ .LogicPackageName}}.{{ .LogicFuncName}}(ctx{{if .PathParam}}, id{{end}}, &req)
	if err != nil {
		gooze.FailWithMessage(ctx, err.Error())
		return
	}
	gooze.Success(ctx, resp){{else}}if err := {{ .LogicPackageName}}.{{ .LogicFuncName}}(ctx{{if .PathParam}}, id{{end}}, &req); err != nil {
		gooze.FailWithMessage(ctx, err.Error())
		return
	}
	gooze.Success(ctx, nil){{end}}{{else}}{{if .ResponseType}}resp, err := {{ .LogicPackageName}}.{{ .LogicFuncName}}(ctx{{if .PathParam}}, id{{end}})
	if err != nil {
		gooze.FailWithMessage(ctx, err.Error())
		return
	}
	gooze.Success(ctx, resp){{else}}if err := {{ .LogicPackageName}}.{{ .LogicFuncName}}(ctx{{if .PathParam}}, id{{end}}); err != nil {
		gooze.FailWithMessage(ctx, err.Error())
		return
	}
	gooze.Success(ctx, nil){{end}}{{end}}
}
`

func (self *generator) GenHandler() (err error) {
	filename := filepath.Join(self.output, self.moduleName, self.handlerPackageName, self.fileName)
	// filename := filepath.Join(self.handlerPackagePath, self.fileName)
	// 删除目标文件（如果存在）
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete existing file: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	for _, service := range self.services {

		if service.Name == "" {
			continue
		}

		if err = self.combineHandlerWrite(service, filename); err != nil {
			return err
		}
	}

	return nil
}

func (self *generator) combineHandlerWrite(service *serviceSpec, filename string) error {
	hasDto := false
	hasRestFul := false
	for _, route := range service.Routes {
		if !hasDto && route.RequestType != "" {
			hasDto = true
		}
		if !hasRestFul && route.RustFulKey != "" {
			hasRestFul = true
		}
	}

	// 如果文件不存在，先写 header
	logicStructName := self.logicName[strings.ToLower(service.Name)]
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		headerData := map[string]interface{}{
			"PackageName":       self.handlerPackageName,
			"HasDto":            hasDto,
			"DtoPackagePath":    filepath.ToSlash(filepath.Join(self.packageName, self.dtoPackagePath)),
			"LogicPackagePath":  filepath.ToSlash(filepath.Join(self.packageName, self.logicPackagePath)),
			"HasRestFul":        hasRestFul,
			"LogicPackageName":  self.logicPackageName,
			"LogicStructName":   logicStructName,
			"LogicInstanceName": gzutil.LcFirst(logicStructName),
		}

		headerStr, err := executeTemplate(handlerHeaderTemplate, headerData)
		if err != nil {
			return err
		}

		if err := writeToFile(filename, headerStr, false); err != nil {
			return err
		}
	}

	for _, route := range service.Routes {
		handlerName := gzutil.UcFirst(service.Name) + route.Name
		handlerData := map[string]interface{}{
			"Summary":          route.Summary,
			"Path":             fmt.Sprintf("/%s/%s", gzutil.SeparateCamel(service.Name, "/"), route.Path),
			"Method":           route.Method,
			"HandlerName":      handlerName,
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"DtoPackageName":   self.dtoPackageName,
			"LogicFuncName":    self.logicFuncName[strings.ToLower(service.Name)+strings.ToLower(route.Name)],
			"LogicPackageName": gzutil.LcFirst(logicStructName),
			"PathParam":        route.RustFulKey,
		}
		contentStr, err := executeHandlerTemplate(handlerContentTemplate, handlerData)
		if err != nil {
			return err
		}

		// 追加 handler
		if err := writeToFile(filename, "\n"+contentStr, true); err != nil {
			return err
		}
	}

	self.formatFileWithGofmt(filename)
	return nil
}

// 渲染模板
func executeHandlerTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("tmpl").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func writeToFile(filename string, content string, append bool) error {
	var f *os.File
	var err error
	if append {
		f, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		f, err = os.Create(filename) // truncate if not append
	}
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
