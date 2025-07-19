package mq

import (
	cfg "filestore-server/config"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

var conn *amqp.Connection
var channel *amqp.Channel

// 如果异常关闭，会接收通知
var notifyClose chan *amqp.Error

func InitMq() {
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !cfg.Config.RabbitMQConfig.AsyncTransferEnable {
		return
	}

	// 初始化notifyClose通道
	notifyClose = make(chan *amqp.Error, 1)

	// 创建连接并创建消息通道
	if initChannel() {
		// 有通道后，监听channel关闭事件
		channel.NotifyClose(notifyClose)
	}

	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				log.Println("onNotifyChannelClosed: 重连中...")
				initChannel()
			}
		}
	}()
}

// 初始化消息channel
func initChannel() bool {
	// 判断是否已经创建
	if channel != nil {
		return true
	}
	// 获得rabbitmq的一个连接
	var err error
	conn, err = amqp.Dial(cfg.Config.RabbitMQConfig.RabbitURL)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return false
	}
	// 打开一个channel，用于消息的发布与接收
	channel, err = conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return false
	}
	return true
}

// 发布消息
func Publish(exchange, routingKey string, msg []byte) bool {
	// 判断channel是否正常
	if !initChannel() {
		return false
	}
	// 发布消息
	err := channel.Publish(
		exchange,
		routingKey,
		false, // 如果没有合适的队列，就会直接丢弃消息，不通知生产者
		// （必须保留参数位置，尽管功能已移除）
		false, // （已弃用）要求立即投递到活跃消费者，否则返回消息。
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		},
	)
	if err != nil {
		fmt.Println("Failed to publish a message:", err)
		return false
	}
	fmt.Println("Publish a message to queue success")
	return true
}
