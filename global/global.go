package global

import (
	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var (
	GLOAB_DB          *gorm.DB            //mysql数据库连接
	GLOAB_DB1         *gorm.DB            //mysql数据库连接
	GLOBAL_REDIS      *redis.Client       //redis连接
	GLOAB_CORN        *cron.Cron          //Cron定时器连接
	GLOBAL_Feishu     *lark.Client        //飞书客户端
	GLOBAL_Kafka_Prod sarama.SyncProducer //kafka生产者
	GLOBAL_Kafka_Cons sarama.Consumer     //kafka消费者
)

// KafMsg 封装kaf消息
func KafMsg(topic, con string, partition int32) *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(con),
		Partition: partition,
	}
}
