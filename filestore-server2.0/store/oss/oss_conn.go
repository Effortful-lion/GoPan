package oss

import (
	cfg "filestore-server/config"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"time"
)

var ossCli *oss.Client

func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	// 修复变量作用域问题，使用赋值操作而不是短声明
	var err error
	ossCli, err = oss.New(cfg.Config.AliyunConfig.OSSEndpoint, cfg.Config.AliyunConfig.OSSAccessKeyID, cfg.Config.AliyunConfig.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

// 获取bucket对象
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(cfg.Config.AliyunConfig.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// 从 oss 下载文件:临时授权下载
func DownloadURL(objectName string) (signedUrl string) {
	bucket := Bucket()
	if bucket == nil {
		fmt.Println("bucket is nil")
		return ""
	}

	expires := time.Now().Add(3600 * time.Second).Unix()
	signedURL, err := bucket.SignURL(objectName, oss.HTTPGet, expires)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedURL
}
