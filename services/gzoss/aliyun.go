package gzoss

import (
	"bytes"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"

	"mime/multipart"
	"sync"
)

type aliYun struct {
	Client *alioss.Client
	Bucket *alioss.Bucket
	once   sync.Once
}

func (self *aliYun) init() {
	self.once.Do(func() {
		accessKey := viper.GetString("Oss.AccessKey")
		secretKey := viper.GetString("Oss.SecretKey")
		endpoint := viper.GetString("Oss.Endpoint")
		bucketName := viper.GetString("Oss.BucketName")
		client, err := alioss.New(endpoint, accessKey, secretKey)
		if err != nil {
			panic(err)
		}

		bucket, err := client.Bucket(bucketName)
		if err != nil {
			panic(err)
		}

		self.Client = client
		self.Bucket = bucket
	})
}

func (self *aliYun) Upload(file *multipart.FileHeader, uploadDir ...string) (*UploadRet, error) {
	savePathUri, filename := getUploadDirAndFilename(file, uploadDir...)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	ossUrl := savePathUri + filename
	err = self.Bucket.PutObject(ossUrl, src)
	if err != nil {
		return nil, err
	}

	return &UploadRet{
		Filename: filename,
		Url:      ossUrl,
	}, err
}

func (self *aliYun) UploadByByte(fileBytes []byte, uploadDir ...string) (*UploadRet, error) {
	savePathUri, filename := getUploadDirAndFilenameByBytes(fileBytes, ".png", uploadDir...)

	reader := bytes.NewReader(fileBytes)
	ossUrl := savePathUri + filename

	err := self.Bucket.PutObject(ossUrl, reader)
	if err != nil {
		return nil, err
	}

	return &UploadRet{
		Filename: filename,
		Url:      ossUrl,
	}, nil
}
