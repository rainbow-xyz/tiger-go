package dbm

/*
// DBHelper Repo层的辅助方法，目前仅用于在usecase做一下db的辅助  ps：目前没有特别优雅的事务处理方式用在本分层骨架中，需要继续探究
type DBHelper struct {
	DB *gorm.DB
}

func NewRepoHelper(db *gorm.DB) *DBHelper {
	return &DBHelper{
		DB: db,
	}
}

func (r DBHelper) BeginTrx() *gorm.DB {
	return r.DB.Begin()
}

func (r DBHelper) Commit() *gorm.DB {
	return r.DB.Commit()
}

func (r DBHelper) Rollback() *gorm.DB {
	return r.DB.Rollback()
}

*/
