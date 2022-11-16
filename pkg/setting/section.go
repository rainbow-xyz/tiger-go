package setting

import "time"

type Config struct {
	ServerCfg   *ServerConfig
	AppCfg      *AppConfig
	DBCfg       *DatabaseConfig
	TenantDBCfg []*DatabaseConfig
	RedisCfg    *RedisConfig
}

type ServerConfig struct {
	RunMode      string
	HttpPort     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type AppConfig struct {
	Name            string
	OAuthTmpToken   string
	DefaultPageSize int
	MaxPageSize     int
	LogSavePath     string
	LogFileExt      string
}

type DatabaseConfig struct {
	ClusterName  string
	DBType       string
	UserName     string
	Password     string
	Host         string
	Port         string
	DBName       string
	TablePrefix  string
	Charset      string
	ParseTime    bool
	MaxIdleConns int
	MaxOpenConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ReadSection 将配置段映射到对应的结构体上
func (s *Setting) ReadSection(k string, v interface{}) error {

	if err := s.vp.UnmarshalKey(k, v); err != nil {
		return err
	}

	return nil
}

func NewConfig(confFile string) (*Config, error) {
	var config Config
	s, err := NewSetting(confFile)
	if err != nil {
		return nil, err
	}

	if err := s.ReadSection("Server", &config.ServerCfg); err != nil {
		return nil, err
	}

	if err := s.ReadSection("App", &config.AppCfg); err != nil {
		return nil, err
	}

	if err := s.ReadSection("Database", &config.DBCfg); err != nil {
		return nil, err
	}

	var tenantDatabaseA DatabaseConfig
	var tenantDatabaseB DatabaseConfig
	if err := s.ReadSection("TenantDatabaseA", &tenantDatabaseA); err != nil {
		return nil, err
	}

	if err := s.ReadSection("TenantDatabaseB", &tenantDatabaseB); err != nil {
		return nil, err
	}

	if err := s.ReadSection("Redis", &config.RedisCfg); err != nil {
		return nil, err
	}

	config.TenantDBCfg = append(config.TenantDBCfg, &tenantDatabaseA)
	config.TenantDBCfg = append(config.TenantDBCfg, &tenantDatabaseB)

	return &config, err
}
