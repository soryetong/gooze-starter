package gzutil

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func AssembleServerPath(filepath string) string {
	return strings.TrimRight(GetServerAddr(), "/") + "/" + strings.TrimLeft(filepath, "/")
}

// 目录是否为空
func DirIsEmpty(dirPath string) (bool, error) {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	return len(dirEntries) == 0, err
}

func FunctionExists(dirPath, funcName string) (bool, error) {
	isEmpty, err := DirIsEmpty(dirPath)
	if err != nil {
		return false, err
	}
	if isEmpty {
		return false, nil
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			if exists, err := parseGoFileForFunction(path, funcName); err == nil && exists {
				return fmt.Errorf("found:%s", path)
			}
		}
		return nil
	})

	if err != nil && strings.HasPrefix(err.Error(), "found:") {
		return true, nil
	}

	return false, err
}

func parseGoFileForFunction(filePath, funcName string) (bool, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filePath, src, parser.AllErrors)
	if err != nil {
		return false, err
	}

	funcMap := make(map[string]bool)
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			funcMap[fn.Name.Name] = true
		}
		return true
	})

	return funcMap[funcName], nil
}

// 文件是否存在
func FileIsExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// 保存文件
func SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func pickPreferredExt(contentType string) string {
	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		return ""
	}

	// 优先选择常用扩展名
	for _, ext := range exts {
		switch ext {
		case
			".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".svg",
			".mp4", ".mov", ".avi", ".wmv", ".flv", ".mkv", ".webm", ".m4v",
			".mp3", ".aac", ".wav", ".ogg", ".flac", ".m4a",
			".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt",
			".zip", ".rar", ".7z", ".tar", ".gz",
			".exe", ".apk", ".dmg", ".sh", ".bin":
			return ext
		}
	}

	// 没有匹配项就返回第一个
	return exts[0]
}

// DownloadFileAutoExt 下载 URL 并保存到指定路径，第二个参数不要包含文件名
// filepath: 文件保存路径, 会自动 md5 原URL 作为新文件的文件名
func DownloadFileAutoExt(url string, filepath string) (string, error) {
	if url == "" || filepath == "" {
		return "", fmt.Errorf("URL is empty")
	}
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	fullPath := filepath + Md5Encode(url) + pickPreferredExt(contentType)
	if err := os.MkdirAll(filepath, 0755); err != nil {
		return "", fmt.Errorf("mkdir failed: %w", err)
	}
	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("file create failed: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("write failed: %w", err)
	}

	return fullPath, nil
}

// SaveBytesToFile 保存字节数组到文件
func SaveBytesToFile(dir, filename string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return fullPath, nil
}
