package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

var cfg Config

type Config struct {
	PgPort              string `yaml:"pgPort" env:"pgPort" env-default:"5432"`
	PgHost              string `yaml:"pgHost" env:"pgHost" env-default:"localhost"`
	PgUser              string `yaml:"pgUser" env:"pgUser" env-default:"pg"`
	PgPassword          string `yaml:"pgPassword" env:"pgPassword" env-default:"pass"`
	DB                  string `yaml:"db" env:"db" env-default:"JITScheduler"`
	Migrate             bool   `yaml:"migrate" env:"migrate" env-default:"true"`
	LogLevel            string `yaml:"logLevel" env:"logLevel" env-default:"info"`
	PORT                int    `yaml:"port" env:"port" env-default:"3000"`
	HOST                string `yaml:"host" env:"port" env-default:"localhost"`
	MaxOpenConnections  int    `yaml:"maxOpenConnections" env:"port" env-default:"5"`
	KafkaBrokerUrl      string `yaml:"kafkaBrokerUrl" env:"kafkaBrokerUrl" env-default:"localhost:9092"`
	RedisHost           string `yaml:"redisHost" env:"redisHost" env-default:"localhost:6379"`
	ReschedulerCronSpec string `yaml:"reschedulerCronSpec" env:"reschedulerCronSpec" env-default:"1s"`
	MonitorCronSpec     string `yaml:"monitorCronSpec" env:"monitorCronSpec" env-default:"5m"`
	KafkaTopic          struct {
		DrainInTopic         string `yaml:"drainInTopic" env:"drainInTopic" env-default:"drain_in_topic"`
		DrainOutTopic        string `yaml:"drainOutTopic" env:"drainOutTopic" env-default:"drain_out_topic"`
		OneSecQueueTopic     string `yaml:"oneSecQueueTopic" env:"oneSecQueueTopic" env-default:"one_sec_queue_topic"`
		FiveSecQueueTopic    string `yaml:"fiveSecQueueTopic" env:"fiveSecQueueTopic" env-default:"five_sec_queue_topic"`
		FifteenSecQueueTopic string `yaml:"fifteenSecQueueTopic" env:"fifteenSecQueueTopic" env-default:"fifteen_sec_queue_topic"`
		OneMinQueueTopic     string `yaml:"oneMinQueueTopic" env:"oneMinQueueTopic" env-default:"one_min_queue_topic"`
		FiveMinQueueTopic    string `yaml:"fiveMinQueueTopic" env:"fiveMinQueueTopic" env-default:"five_min_queue_topic"`
		FifteenMinQueueTopic string `yaml:"fifteenMinQueueTopic" env:"fifteenMinQueueTopic" env-default:"fifteen_min_queue_topic"`
		OneHourQueueTopic    string `yaml:"oneHourQueueTopic" env:"oneHourQueueTopic" env-default:"one_hour_queue_topic"`
		ConsumerGroupId      string `yaml:"consumerGroupId" env:"consumerGroupId" env-default:"jit_scheduler_consumer_group"`
	}
}

func InitConfig() error {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return err
	}
	return nil
}

func SetConfig(newConfig Config) {
	cfg = newConfig
}

func GetConfig() Config {
	return cfg
}
