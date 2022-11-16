package core

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

//type CtxKey string

type BitInt int

// Value 应对向数据库中写入bit(1)的问题
func (b BitInt) Value() (driver.Value, error) {
	if b == 0 {
		return []byte{0}, nil
	} else {
		return []byte{1}, nil
	}
}

// Scan 应对从数据库中读取bit(1)的问题
func (b *BitInt) Scan(src interface{}) error {
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}

	if v[0] == 0 {
		*b = 0
	} else {
		*b = 1
	}
	return nil
}

// MarshalBinary 应对go-redis 解析编码BitInt的问题
func (b BitInt) MarshalBinary() (data []byte, err error) {
	return json.Marshal(b)
}
