package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
	"log"

	cfg "filestore-server/config"
)

var cephConn *s3.S3

// GetCephConnection : 获取ceph连接
func GetCephConnection() *s3.S3 {
	if cephConn != nil {
		return cephConn
	}
	// 1. 初始化ceph的一些信息

	auth := aws.Auth{
		AccessKey: cfg.Config.CephConfig.CephAccessKey,
		SecretKey: cfg.Config.CephConfig.CephSecretKey,
	}

	curRegion := aws.Region{
		Name:                 "default",
		EC2Endpoint:          cfg.Config.CephConfig.CephGWEndpoint,
		S3Endpoint:           cfg.Config.CephConfig.CephGWEndpoint,
		S3BucketEndpoint:     "",
		S3LocationConstraint: false,
		S3LowercaseBucket:    false,
		Sign:                 aws.SignV2,
	}

	// 2. 创建S3类型的连接
	return s3.New(auth, curRegion)
}

// GetCephBucket : 获取指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	b := conn.Bucket(bucket)
	// 检查存储桶是否存在
	_, err := b.List("", "", "", 0)
	if err != nil {
		if s3err, ok := err.(*s3.Error); ok && s3err.StatusCode == 404 {
			log.Printf("存储桶 %s 不存在，尝试创建...", bucket)
			// 创建存储桶
			err = b.PutBucket(s3.PublicRead)
			if err != nil {
				log.Printf("创建存储桶 %s 失败: %v", bucket, err)
				return nil
			}
			log.Printf("存储桶 %s 创建成功", bucket)
		} else {
			log.Printf("访问存储桶 %s 出错: %v", bucket, err)
			return nil
		}
	}
	return b
}

// PutObject : 上传文件到ceph集群
func PutObject(bucket string, path string, data []byte) error {
	buck := GetCephBucket(bucket)
	return buck.Put(path, data, "octet-stream", s3.PublicRead)
}
