package gzoss

import (
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

type local struct{}

func (*local) Upload(file *multipart.FileHeader, uploadDir ...string) (*UploadRet, error) {
	// 获取上传目录和文件名
	savePathUri, filename := getUploadDirAndFilename(file, uploadDir...)

	// 检查并创建上传目录
	isPath, mkdirErr := gzutil.FileIsExist(savePathUri)
	if mkdirErr != nil {
		return nil, mkdirErr
	}
	if !isPath {
		_ = os.MkdirAll(savePathUri, os.ModePerm)
	}

	// 保存文件
	filePath := filepath.Join(savePathUri, filename)
	if err := gzutil.SaveFile(file, filePath); err != nil {
		return nil, err
	}

	return &UploadRet{
		Hash:     gzutil.Md5Encode(filePath),
		Filename: filename,
		Url:      filePath,
	}, nil
}

func (self *local) UploadByByte(fileBytes []byte, uploadDir ...string) (*UploadRet, error) {
	savePathUri, filename := getUploadDirAndFilenameByBytes(fileBytes, ".png", uploadDir...)
	if err := os.MkdirAll(savePathUri, 0755); err != nil {
		return nil, err
	}

	fullPath := savePathUri + filename
	err := os.WriteFile(fullPath, fileBytes, 0644)
	if err != nil {
		return nil, err
	}
	return &UploadRet{
		Filename: filename,
		Url:      fullPath,
	}, nil
}
