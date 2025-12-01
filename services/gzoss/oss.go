package gzoss

import (
	"fmt"
	"math/rand"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/soryetong/gooze-starter/pkg/gzutil"
	"github.com/spf13/viper"
)

type OssType string

const (
	OssTypeLocal  OssType = "local"
	OssTypeQiNiu  OssType = "qiniu"
	OssTypeAliYun OssType = "aliyun"
)

type UploadRet struct {
	Filename string `json:"filename"`
	Hash     string `json:"hash"`
	Url      string `json:"url"`
	Key      string `json:"key"`
}

type oss interface {
	Upload(file *multipart.FileHeader, uploadDir ...string) (*UploadRet, error)
	UploadByByte(fileBytes []byte, uploadDir ...string) (*UploadRet, error)
}

func New(ossType OssType) oss {
	return start(ossType)
}

func NewAliYun() oss {
	return start(OssTypeAliYun)
}

func NewQiNiu() oss {
	return start(OssTypeQiNiu)
}

func NewLocal() oss {
	return start(OssTypeLocal)
}

func NewByConf() oss {
	return start(OssType(viper.GetString("Oss.Type")))
}

func start(ossType OssType) oss {
	switch ossType {
	case OssTypeAliYun:
		c := &aliYun{}
		c.init()
		return c
	case OssTypeQiNiu:
		return &qiNiu{}
	default:
		return &local{}
	}
}

func getUploadDirAndFilename(fileHeader *multipart.FileHeader, uploadDir ...string) (string, string) {
	version := time.Now().Format("2006/01/02")
	var dir string
	if len(uploadDir) > 0 {
		dir = uploadDir[0]
	} else {
		dir = viper.GetString("Oss.SavePath")
	}
	if dir == "" {
		dir = "./static/storage/attach/"
	}
	dir = strings.TrimRight(dir, "/") + "/"

	// 读取文件后缀
	ext := path.Ext(fileHeader.Filename)
	// 读取文件名并加密
	name := gzutil.Md5Encode(fileHeader.Filename + fmt.Sprintf("%d%04d", time.Now().Unix(), rand.Int31()))
	// 拼接新文件名
	filename := name + ext

	return dir + version + "/", filename
}

func getUploadDirAndFilenameByBytes(fileBytes []byte, ext string, uploadDir ...string) (string, string) {
	version := time.Now().Format("2006/01/02")
	var dir string
	if len(uploadDir) > 0 {
		dir = uploadDir[0]
	} else {
		dir = viper.GetString("Oss.SavePath")
	}
	if dir == "" {
		dir = "./static/storage/attach/"
	}
	dir = strings.TrimRight(dir, "/") + "/"

	// 使用 MD5 + 时间戳 + 随机数生成唯一文件名
	name := gzutil.Md5Encode(fmt.Sprintf("%d%04d", time.Now().Unix(), rand.Int31()))
	filename := name + ext

	return dir + version + "/", filename
}
