package config

const (
	// AsyncTransferEnable : 是否开启文件异步转移
	AsyncTransferEnable = true
	// RabbitURL: rabbitmq服务的入口url
	RabbitURL = "amqp://guest:guest@localhost:5672/"
	// TransExchangeName: 转移文件的交换机
	TransExchangeName = "uploadserver.trans"
	// TransOSSQueueName: oss转移队列
	TransOSSQueueName = "uploadserver.trans.oss"
	// TransOSSErrQueueName: oss转移出错队列(存储出错的任务)
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	// TransOSSRoutingKey: oss转移绑定的routingkey(路由键)
	TransOSSRoutingKey = "oss"
)
