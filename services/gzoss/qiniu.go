package gzoss

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
	"github.com/spf13/viper"
)

type qiNiu struct{}

func (*qiNiu) Upload(file *multipart.FileHeader, uploadDir ...string) (*UploadRet, error) {
	// 获取上传目录和文件名
	savePathUri, filename := getUploadDirAndFilename(file, uploadDir...)

	accessKey := viper.GetString("Oss.AccessKey")
	secretKey := viper.GetString("Oss.SecretKey")
	bucket := viper.GetString("Oss.BucketName")
	ossUrl := viper.GetString("Oss.Url")
	if accessKey == "" || secretKey == "" || ossUrl == "" || bucket == "" {
		return nil, errors.New("config has empty value")
	}
	mac := credentials.NewCredentials(accessKey, secretKey)
	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})

	ret := new(UploadRet)
	objectName := savePathUri + filename
	fileRet, err := file.Open()
	if err != nil {
		return nil, err
	}
	err = uploadManager.UploadReader(context.Background(), fileRet, &uploader.ObjectOptions{
		BucketName: bucket,
		ObjectName: &objectName,
		FileName:   filename,
	}, &ret)

	return &UploadRet{
		Hash:     ret.Hash,
		Filename: filename,
		Url:      objectName,
	}, err
}

func (self *qiNiu) UploadByByte(fileBytes []byte, uploadDir ...string) (*UploadRet, error) {
	return nil, nil
}
