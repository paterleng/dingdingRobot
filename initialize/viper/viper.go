package viper

import (
	"ding/utils"
	"fmt"
	"github.com/fsnotify/fsnotify"

	//"github.com/fsnotify/fsnotify"

	"github.com/spf13/viper"
)

var Conf = new(AppConfig) //这是一个指针，全局变量，用来保存程序的所有配置信息
type AppConfig struct {
	*App         `mapstructure:"app"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
	*LogConfig   `mapstructure:"log"`
	*Auth        `mapstructure:"auth"`
	*KafkaConfig `mapstructure:"kafka"`
}
type Auth struct {
	Jwt_Expire int `mapstructure:"jwt_expire"`
}
type App struct {
	Name      string `mapstructure:"name"`
	Mode      string `mapstructure:"mode"`
	Version   string `mapstructure:"version"`
	Port      int    `mapstructure:"port"`
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
}
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}
type RedisConfig struct {
	DB       int    `mapstructure:"db"`
	Addr     string `mapstructure:"addr"`
	Password string `mapsturcture:"password"`
}
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type KafkaConfig struct {
	Address string `mapstructure:"address"`
}

func Init() (err error) {
	var config string
	config = utils.ConfigFile
	viper.SetConfigFile(config)
	//方式一：
	viper.SetConfigFile(config) // 只用这一句即可
	//方式二：
	//viper.SetConfigName("config") // 指定配置文件(不需要加后缀)
	//viper.AddConfigPath(".")      //指定查找配置文件的路径（相对路径）
	//viper.AddConfigPath("./conf") //指定查找配置文件的路径（相对路径）
	err = viper.ReadInConfig()
	if err != nil {
		//读取配置信息失败了
		fmt.Println("viper.ReadInConfig() failed,err:#{err}\n")
		return
	}
	//把读取到的配置信息给反序列化到结构体中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("vip Unmarshal failed,err:%v\n", err)
	}
	//viper热加载设置
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了....")
		if err := viper.Unmarshal(Conf); err != nil { //重新反序列化到结构体
			fmt.Printf("viper Unmarshal failed,err:%v\n", err)
		}
	})
	return
}
