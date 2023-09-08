package initialize

import (
	"ding/global"
	"ding/initialize/viper"
	"fmt"
	"github.com/Shopify/sarama"
)

/**
*
* @author yth
* @language go
* @since 2023/6/10 21:14
 */

func KafkaInit() (err error) {
	config := sarama.NewConfig()
	// 用于指示生产者在成功发送消息后是否返回成功的响应。
	config.Producer.Return.Successes = true
	// Kafka集群的地址
	brokers := []string{viper.Conf.KafkaConfig.Address}
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	// defer func() { _ = client.Close() }()

	// 创建一个Kafka生产者
	p, err := sarama.NewSyncProducerFromClient(client)
	global.GLOBAL_Kafka_Prod = p
	if err != nil {
		fmt.Println(err)
		return
	}
	// defer func() { _ = producer.Close() }()

	// 创建一个Kafka消费者
	c, err := sarama.NewConsumerFromClient(client)
	global.GLOBAL_Kafka_Cons = c
	if err != nil {
		fmt.Println(err)
		return
	}
	// defer func() { _ = consumer.Close() }()
	return nil
}
