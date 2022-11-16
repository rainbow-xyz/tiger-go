package dbm

import (
	"fmt"
	mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/url"
	"saas_service/pkg/setting"
	"saas_service/pkg/xlog"
)

type XDB struct {
	CoreDB    *gorm.DB
	TenantDBS map[string]*gorm.DB
}

var xdb *XDB

const TenantDBInsA = "TenantDBInsA"
const TenantDBInsB = "TenantDBInsB"

func (tdb *XDB) GetTenantDBIInsConn(TenantID string) (*gorm.DB, error) {
	// 获取租户id

	a, _ := xdb.TenantDBS["TenantDatabaseA"]
	b, _ := xdb.TenantDBS["TenantDatabaseB"]
	// 查询租户所在集群
	insMap := make(map[string]*gorm.DB)
	insMap["656"] = a
	insMap["672"] = b

	// 返回db实例
	return a, nil
}

func ConnectDB(dbCfg *setting.DatabaseConfig) (*gorm.DB, error) {

	logger := xlog.NewZGLogger(xlog.Logger)
	//logger.SetAsDefault() // optional: configure gorm to use this zapgorm.Logger for callbacks
	//logger.SetAsDefault() // optional: configure gorm to use this zapgorm.Logger for callbacks

	s := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbCfg.UserName, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Shanghai")
	dsn := fmt.Sprintf("%s?%s", s, val.Encode())
	dbConn, err := gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{

			Logger: logger,
		},
	)
	if err != nil {
		return nil, err
	}
	return dbConn, err
}

func ConnectXDB(coreDBCfg *setting.DatabaseConfig, tenantDBCfg []*setting.DatabaseConfig) (*XDB, error) {
	if xdb != nil {
		return xdb, nil
	}

	var _xdb XDB = XDB{
		CoreDB:    nil,
		TenantDBS: map[string]*gorm.DB{},
	}
	_db, err := ConnectDB(coreDBCfg)
	if err != nil {
		return nil, err
	}
	_xdb.CoreDB = _db

	for _, v := range tenantDBCfg {
		_db, err := ConnectDB(v)
		if err != nil {
			return nil, err
		}

		_xdb.TenantDBS[v.ClusterName] = _db
	}
	xdb = &_xdb
	return &_xdb, nil
}

/*
type DatabaseCfg struct {
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

func ConnectDB(cfgFile string) (*sql.DB, error) {
	// 获取db配置
	dbCfg, err := loadDBCfg(cfgFile)
	if err != nil {
		panic(err)
	}

	// 连接db
	s := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbCfg.UserName, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Shanghai")
	dsn := fmt.Sprintf("%s?%s", s, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)
	if err != nil {
		return nil, err
	}

	err = dbConn.Ping()
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

func ConnectGormDB(cfgFile string) (*gorm.DB, error) {

	sqlDB, err := ConnectDB(cfgFile)
	if err != nil {
		return nil, err
	}
	dbConn, err := gorm.Open(gmysql.New(gmysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	return dbConn, err
}

func loadDBCfg(cfgFile string) (*DatabaseCfg, error) {
	vp := viper.New()
	vp.SetConfigFile(cfgFile)

	if err := vp.ReadInConfig(); err != nil {
		panic(err)
	}

	var dbcfg *DatabaseCfg

	if err := vp.UnmarshalKey("Database", &dbcfg); err != nil {
		panic(err)
	}

	return dbcfg, nil
}
*/
