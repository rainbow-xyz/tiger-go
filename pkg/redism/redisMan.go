package redism

import (
	"github.com/go-redis/redis/v8"
	"saas_service/pkg/setting"
	"strconv"
)

// RedisOptions defines options for redis cluster.
type RedisOptions struct {
	Host     string `json:"host" mapstructure:"host"                     description:"Redis service host address"`
	Port     int    `json:"port" mapstructure:"port"`
	Password string `json:"password" mapstructure:"password"`
	DB       int    `json:"port" mapstructure:"db"`
}

var redisClientSingleton *redis.Client

func ConnectRedis(redisCfg *setting.RedisConfig) (*redis.Client, error) {

	if redisClientSingleton != nil {
		return redisClientSingleton, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Host + ":" + strconv.Itoa(redisCfg.Port),
		Password: redisCfg.Password, // no password set
		DB:       redisCfg.DB,       // use default DB
	})

	redisClientSingleton = client

	return client, nil
}

/*

func ConnectRedis(cfgFile string) (*redis.Client, error) {

	if redisClientSingleton != nil {
		return redisClientSingleton, nil
	}

	rdbCfg, err := loadRedisCfg(cfgFile)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     rdbCfg.Host + ":" + strconv.Itoa(rdbCfg.Port),
		Password: rdbCfg.Password, // no password set
		DB:       rdbCfg.DB,       // use default DB
	})

	redisClientSingleton = client

	return client, nil
}

func loadRedisCfg(cfgFile string) (*RedisOptions, error) {
	vp := viper.New()
	vp.SetConfigFile(cfgFile)

	if err := vp.ReadInConfig(); err != nil {
		panic(err)
	}

	var rdbCfg *RedisOptions

	if err := vp.UnmarshalKey("Redis", &rdbCfg); err != nil {
		panic(err)
	}

	return rdbCfg, nil
}

*/
