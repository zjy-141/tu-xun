package model

import (
	"fmt"
	"log"
	"time"
	dblog "tu-xun/logger"

	"tu-xun/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local",
		config.Config.MysqlUser,
		config.Config.MysqlPass,
		config.Config.MysqlHost,
		config.Config.MysqlPort,
		config.Config.MysqlName)
	var dbLogger logger.Interface
	if dblog.DatabaseLogger == nil {
		dbLogger = logger.Default.LogMode(logger.Info)
	} else {
		logLevels := map[string]int{
			"error": 2,
			"warn":  3,
			"info":  4,
		}

		levels, ok := logLevels[config.Config.LogLevel]
		if !ok {
			levels = 4
		}
		dbLogger = logger.New(
			log.New(dblog.DataLogger{Logger: dblog.DatabaseLogger}, "\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.LogLevel(levels),
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dbLogger})
	if err != nil {
		panic(err)
	}

	DB = db

	if !config.Config.AppProd {
		initModel()
	}
}

// initModel 初始化 GORM 数据库连接并执行自动迁移
func initModel() {

	// example
	// begin
	DB.AutoMigrate(
		&User{},
		&Activity{}, &AttemptRewardTier{},
		&Photo{},
		&Attempt{},
		&Good{},
		&Exchange{},
		&Comment{},
		&Message{},
		&Notice{},
		&Feedback{}, &FeedbackMedia{},
		&Like{},
		&ScoreLog{})
	//end
	// 默认初始化一批参数
	initData()
}

func initData() {

	if err := DB.FirstOrCreate(&User{
		BaseModel: BaseModel{ID: 1},
		NetID:     "1",
		Name:      "系统",
		Nickname:  "tz",
		Level:     1,
	}).Error; err != nil {
		panic(err)
	}

	// if err := DB.FirstOrCreate(&User{
	// 	BaseModel: BaseModel{ID: 2},
	// 	NetID:     "2251416412",
	// 	Name:      "张继尧",
	// 	Nickname:  "J",
	// 	Level:     4,
	// }).Error; err != nil {
	// 	panic(err)
	// }
}
