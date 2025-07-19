// config.go
// 将配置文件加载到内存中，方便使用
package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

var Config = &AppConfig{}

// app应用配置
type AppConfig struct {
	MysqlConfig    MysqlConfig    `mapstructure:"MYSQL"`
	RedisConfig    RedisConfig    `mapstructure:"Redis"` // 改为 Redis
	GinConfig      GinConfig      `mapstructure:"Gin"`   // 改为 Gin
	EtcdConfig     EtcdConfig     `mapstructure:"Etcd"`
	ServiceConfig  ServiceConfig  `mapstructure:"Service"`
	RabbitMQConfig RabbitMQConfig `mapstructure:"RabbitMQ"`
	AliyunConfig   AliyunConfig   `mapstructure:"AliYun"` // 改为 AliYun
	Domain         Domain         `mapstructure:"Domain"`
	CephConfig     CephConfig     `mapstructure:"Ceph"`
}

type CephConfig struct {
	CephAccessKey  string `mapstructure:"CephAccessKey"`
	CephSecretKey  string `mapstructure:"CephSecretKey"`
	CephGWEndpoint string `mapstructure:"CephGWEndpoint"`
	CephBucket     string `mapstructure:"CephBucket"`
}

// mysql配置
type MysqlConfig struct {
	DBHost     string `mapstructure:"DBHost"`
	DBPort     int    `mapstructure:"DBPort"` // 类型改为 int，匹配 YAML
	DBUser     string `mapstructure:"DBUser"`
	DBPassword string `mapstructure:"DBPassWord"` // 改为 DBPassWord
	DBName     string `mapstructure:"DBName"`
	DBCharset  string `mapstructure:"Charset"` // 改为 Charset
}

// redis配置
type RedisConfig struct {
	RedisHost     string `mapstructure:"RedisHost"` // 改为 RedisHost
	RedisPort     int    `mapstructure:"RedisPort"` // 改为 RedisPort，类型 int
	RedisPassword string `mapstructure:"RedisPassword"`
}

// gin配置
type GinConfig struct {
	AppMode string `mapstructure:"AppMode"`
	AppHost string `mapstructure:"HttpHost"` // 改为 HttpHost
	AppPort string `mapstructure:"HttpPort"` // 改为 HttpPort
}

// etcd配置
type EtcdConfig struct {
	EtcdHost string `mapstructure:"EtcdHost"`
	EtcdPort int    `mapstructure:"EtcdPort"` // 类型 int
}

// rabbitmq配置
type RabbitMQConfig struct {
	RabbitMQ         string `mapstructure:"RabbitMQ"`
	RabbitMQHost     string `mapstructure:"RabbitMQHost"`
	RabbitMQPort     int    `mapstructure:"RabbitMQPort"` // 类型 int
	RabbitMQUser     string `mapstructure:"RabbitMQUser"`
	RabbitMQPassword string `mapstructure:"RabbitMQPassWord"` // 改为 RabbitMQPassWord
	// 下面是关于操作的配置
	//  # AsyncTransferEnable : 是否开启文件异步转移
	AsyncTransferEnable bool `mapstructure:"AsyncTransferEnable"`
	//  # RabbitURL: rabbitmq服务的入口url
	RabbitURL string `mapstructure:"RabbitURL"`
	//  # TransExchangeName: 转移文件的交换机
	TransExchangeName string `mapstructure:"TransExchangeName"`
	//  # TransOSSQueueName: oss转移队列
	TransOSSQueueName string `mapstructure:"TransOSSQueueName"`
	//  # TransOSSErrQueueName: oss转移出错队列(存储出错的任务)
	TransOSSErrQueueName string `mapstructure:"TransOSSErrQueueName"`
	//  # TransOSSRoutingKey: oss转移绑定的routingkey(路由键)
	TransOSSRoutingKey string `mapstructure:"TransOSSRoutingKey"`
}

// aliyun配置
type AliyunConfig struct {
	OSSBucket          string `mapstructure:"OSSBucket"` // 改为 Bucket
	OSSAccessKeyID     string `mapstructure:"OSSAccessKeyID"`
	OSSAccessKeySecret string `mapstructure:"OSSAccessKeySecret"`
	OSSEndpoint        string `mapstructure:"OSSEndpoint"`
}

// 服务配置(记录多个服务节点)
type ServiceConfig struct {
	DBProxyServiceAddress  string `mapstructure:"DBProxyServiceAddress"`
	DownloadServiceAddress string `mapstructure:"DownloadServiceAddress"`
	UploadServiceAddress   string `mapstructure:"UploadServiceAddress"`
	TransferServiceAddress string `mapstructure:"TransferServiceAddress"`
	UserServiceAddress     string `mapstructure:"UserServiceAddress"`
}

type Domain struct {
	DBProxyServiceDomain  string `mapstructure:"DBProxyServiceDomain"`
	DownloadServiceDomain string `mapstructure:"DownloadServiceDomain"`
	UploadServiceDomain   string `mapstructure:"UploadServiceDomain"`
	TransferServiceDomain string `mapstructure:"TransferServiceDomain"`
	UserServiceDomain     string `mapstructure:"UserServiceDomain"`
}

func InitConfig() {
	// 使用 filepath.Join 自动处理路径分隔符
	//viper.SetConfigName("config")
	//viper.SetConfigType("yaml")
	//viper.AddConfigPath("./config")
	configFilePath := "D:/GolandCode/src/filestore-server/config/config.yaml"
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&Config); err != nil {
		panic(err)
	}
	// 监控配置文件，当发生变化时，重新加载配置文件
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		if err := viper.Unmarshal(&Config); err != nil {
			panic(err)
		}
		log.Println("配置文件已更新")
	})
	log.Println("配置文件加载成功")
}

// TODO 有时间做：从etcd获取配置文件
// func InitConfig() (err error) {
// 	// 从etcd获取配置文件
// 	if err = discovery.GetConfig("config.yaml"); err != nil {}
// 	// 读取配置文件

// 	// 将配置文件中的内容映射到结构体中

// 	return nil
// }
