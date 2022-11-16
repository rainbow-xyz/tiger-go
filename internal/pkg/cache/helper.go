package cache

import (
	"encoding/json"
	"github.com/marmotedu/errors"
	"saas_service/internal/pkg/code"
)

func Encode(val interface{}) (string, error) {
	uBytes, err := json.Marshal(val)
	if err != nil {
		return "", errors.WithCode(code.ErrEncodingJSON, err.Error())
	}
	return string(uBytes), nil
}

func Decode(val string, objV interface{}) error {
	err := json.Unmarshal([]byte(val), &objV)
	if err != nil {
		return errors.WithCode(code.ErrDecodingJSON, err.Error())
	}
	return nil
}
