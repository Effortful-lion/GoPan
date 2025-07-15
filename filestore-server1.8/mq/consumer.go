package mq

import (
	cfg "filestore-server/config"
	"fmt"
	"log"
)

// 消费者逻辑
// 循环监听队列，消息到达后，进行处理；否则处理

// 定义一个channel，用于控制
var done chan bool

// 开始消费：队列名、消费者名称、回调函数
func StartConsume(qName, cName string, callback func(msg []byte) bool) {

	// 新增交换器声明
	err := channel.ExchangeDeclare(
		cfg.TransExchangeName, // 使用配置中的交换器名称
		"direct",              // 交换器类型
		true,                  // 持久化
		false,                 // 自动删除
		false,                 // 内部交换器
		false,                 // 不等待
		nil,
	)
	if err != nil {
		log.Printf("声明交换器失败: %v", err)
		return
	}

	// 新增队列声明
	_, err = channel.QueueDeclare(
		qName,
		true,  // 持久化
		false, // 自动删除
		false, // 排他性队列
		false, // 等待服务器响应
		nil,
	)
	if err != nil {
		fmt.Printf("声明队列失败: %v\n", err)
		return
	}

	// 绑定队列到默认交换器（根据实际配置调整）
	err = channel.QueueBind(
		qName,
		cfg.TransOSSRoutingKey, // 使用配置中的路由键
		cfg.TransExchangeName,  // 使用配置中的交换器
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("绑定队列失败: %v\n", err)
		return
	}

	// 1. 通过channel.Consume获得消息信道
	msgchannel, err := channel.Consume(
		qName,
		cName,
		true,  // 自动回复 ack
		false, // 非唯一的消费者（true是唯一的消费者）
		false, // rabbitmq只能设置为false
		false, // no-local RabbitMQ 会确认消费请求，并返回一个 ConsumerTag。
		nil,   // args
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 启动一个协程处理消息, 防止阻塞后面的代码执行

	done = make(chan bool)

	go func() {
		// 2. 循环获取队列的消息
		for msg := range msgchannel {
			// 3. 调用callback方法，处理消息
			processSuc := callback(msg.Body)
			if !processSuc {
				// TODO : 将任务写到对应的错误队列，用于异常情况的重试
			}
		}
	}()

	// 如果没有消息，将一直阻塞，等待任务完成
	<-done

	// 4. 关闭channel
	channel.Close()

}

// StopConsume : 停止监听队列
func StopConsume() {
	done <- true
}
