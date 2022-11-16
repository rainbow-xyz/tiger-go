package core

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/segmentio/ksuid"
	"strings"
)

func GetTableFilterFieldsFromCtx(ctx context.Context, tableName string) string {
	var fields string
	val := ctx.Value("__" + strings.ToUpper(tableName) + ":FIELDS")
	if val != nil {
		fields = val.(string)
	}

	if fields == "" {
		return "---INVALID FIELDS---"
	}

	return strings.TrimSpace(fields)
}

// SetTableFilterFieldsToCtx 设置创建或更新时过滤字段
func SetTableFilterFieldsToCtx(ctx context.Context, tableName string, fields string) context.Context {
	return context.WithValue(ctx, "__"+strings.ToUpper(tableName)+":FIELDS", fields)
}

// GetTableCUFilterFieldsFromCtx 获取创建或更新时过滤字段
func GetTableCUFilterFieldsFromCtx(ctx context.Context, tableName string) []string {
	var fields string
	val := ctx.Value("__CU" + strings.ToUpper(tableName) + ":FIELDS")
	if val != nil {
		fields = val.(string)
	}

	fields = strings.TrimSpace(fields)
	if fields == "" || fields == "*" {
		return []string{"---INVALID FIELDS---", "---INVALID FIELDS---"}
	}

	filter := strings.Split(fields, ",")
	j := 2 - len(filter)
	if j < 0 {
		j = 0
	}
	for i := 0; i < j; i++ {
		filter = append([]string{"---INVALID FIELDS---"}, filter...)
	}

	return filter
}

func SetTableCUFilterFieldsToCtx(ctx context.Context, tableName string, fields string) context.Context {
	return context.WithValue(ctx, "__CU"+strings.ToUpper(tableName)+":FIELDS", fields)
}

func GenerateAccessToken() string {
	h := md5.New()
	h.Write([]byte(ksuid.New().String()))
	return hex.EncodeToString(h.Sum(nil))
}
