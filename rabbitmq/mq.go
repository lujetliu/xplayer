package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
) //导入mq包

// TODO: 从配置文件读取
// MQURL 格式 amqp://账号：密码@rabbitmq服务器地址：端口号/vhost (默认是5672端口)
// 端口可在 /etc/rabbitmq/rabbitmq-env.conf 配置文件设置，也可以启动后通过netstat -tlnp查看
const MQURL = "amqp://guest:guest@127.0.0.1:5672/"

var Rmq *rabbitMQ

func init() {
	rmq, err := newRabbitMQ("video", "video-channel", "video-image")
	if err != nil {
		panic(err)
	}
	Rmq = rmq
}

type rabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机
	Exchange string
	// routing Key
	RoutingKey string
	//MQ链接字符串
	Mqurl string
}

// 创建结构体实例
func newRabbitMQ(queueName, exchange, routingKey string) (*rabbitMQ, error) {
	rabbitMQ := rabbitMQ{
		QueueName:  queueName,
		Exchange:   exchange,
		RoutingKey: routingKey,
		Mqurl:      MQURL,
	}
	var err error
	//创建rabbitmq连接
	rabbitMQ.Conn, err = amqp.Dial(rabbitMQ.Mqurl)
	if err != nil {
		log.Println("创建连接失败", err.Error())
		return &rabbitMQ, err
	}

	//创建Channel
	rabbitMQ.Channel, err = rabbitMQ.Conn.Channel()
	if err != nil {
		log.Println("创建channel失败", err.Error())
		return &rabbitMQ, err
	}

	return &rabbitMQ, nil
}

// 释放资源,建议NewRabbitMQ获取实例后 配合defer使用
func (mq *rabbitMQ) ReleaseRes() {
	mq.Conn.Close()
	mq.Channel.Close()
}
