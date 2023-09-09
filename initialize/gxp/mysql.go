package gxp

import (
	"ding/global"
	"ding/initialize/viper"
	"ding/model/dingding"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func GxpInit(cfg *viper.MySQLConfig) (err error) {
	DSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/gxp?charset=utf8&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		//DSN: "root:123456@tcp(121.43.119.224:3306)/gorm_class?charset=utf8mb4&parseTime=True&loc=Local",
		DSN: DSN, // 1. 连接信息
	}), &gorm.Config{ // 2. 选项
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true, //不用物理外键，使用逻辑外键
	})
	if err != nil {
		zap.L().Debug("数据库链接失败")
		return err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10) //
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	global.GLOAB_DB1 = db
	if err != nil {
		fmt.Println(err)
	}

	global.GLOAB_DB1.AutoMigrate(&dingding.TongXinUser{}, &dingding.Record{})
	return nil
}
func Init() (err error) {
	err = GxpInit(viper.Conf.MySQLConfig)
	if err != nil {
		fmt.Printf("init mysql failed ,err:%v\n", err)
		zap.L().Error(fmt.Sprintf("init mysql failed ,err:%v\n", err))
		return
	}
	err = CronSendOne()
	if err != nil {
		zap.L().Error("关鑫鹏22：00定时任务发送失败，", zap.Error(err))
	} //晚上10点的定时提醒
	err = CronSendTwo()
	if err != nil {
		zap.L().Error("关鑫鹏22：20定时任务发送失败，", zap.Error(err))
	} //晚上10:20@未到宿舍的人员
	err = CronSendThree()
	if err != nil {
		zap.L().Error("关鑫鹏22：30定时任务发送失败，", zap.Error(err))
	} //晚上10：35统计结果发给gxp
	return
}

//x := dingding.TongXinUser{}
//err = x.ImportUserToMysql()
//if err != nil {
//fmt.Println("导入人员失败")
//}
