package config

import (
	"gorm.io/gorm/logger"
	"time"
)

type Config struct {
	Http   *Http          `yaml:"http" validate:"required"`
	Mysql  *Mysql         `yaml:"mysql" validate:"required"`
	Redis  *Redis         `yaml:"redis" validate:"required"`
	Pubsub *Pubsub        `yaml:"pubsub" validate:"required"`
	Alert  *Alert         `yaml:"alert" validate:"required"`
	Logger *logger.Config `yaml:"logger" validate:"required"`
}
type Http struct {
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}
type Redis struct {
	URL string `yaml:"url" validate:"required"`
}
type Mysql struct {
	DSN string `yaml:"dsn" validate:"required"`
}

type Alert struct {
	Type   string       `yaml:"type" validate:"required,oneof=feishu noop"`
	Feishu *AlertFeishu `yaml:"feishu"`
}
type AlertFeishu struct {
	WebhookURL string `yaml:"webhook_url" validate:"required"`
	SignSecret string `yaml:"sign_secret" validate:"required"`
}
type Pubsub struct {
	MysqlDSN string `yaml:"mysql_dsn" validate:"required"`
}
