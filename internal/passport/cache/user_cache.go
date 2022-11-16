package cache

import (
	"context"
	"github.com/fatih/structs"
	"github.com/go-redis/redis/v8"
	"github.com/marmotedu/errors"
	"saas_service/internal/pkg/code"
	"saas_service/internal/pkg/domain/passport"
	"saas_service/pkg/core"
	"strconv"
	"strings"
	"time"
)

type userCache struct {
	Redis    *redis.Client
	UserRepo passport.UserRepo
}

const UserPrefix = "XPASSPORT:USER"
const UserTTL = 86400 * 30 * time.Second         // user通用
const UserTokenTTL = 86400 * 30 * time.Second    // token的时间
const UserLoginListTTL = 86400 * 2 * time.Second // 每日登录用户id列表有效时间

func NewUserCache(rdb *redis.Client, u passport.UserRepo) passport.UserCache {
	return &userCache{
		Redis:    rdb,
		UserRepo: u,
	}
}

func (u userCache) GetCacheKeyUID(userID int64) string {
	return UserPrefix + ":UID:" + strconv.FormatInt(userID, 10)
}

func (u userCache) GetCacheKeyToken(accessToken string) string {
	return UserPrefix + ":ACCESS_TOKEN2ID:" + accessToken
}

func (u userCache) GetCacheKeyMisToken(accessToken string) string {
	return UserPrefix + ":MIS_ACCESS_TOKEN2ID:" + accessToken
}

func (u userCache) GetCacheKeyLoginUsersByDate(date string) string {
	return UserPrefix + ":LOGIN_USERS:" + date
}

func (u userCache) GetCacheKeyMisUserSession(userID int64) string {
	return UserPrefix + ":MIS_UID_SESSION:" + strconv.FormatInt(userID, 10)
}

func (u userCache) GetUserByID(ctx context.Context, userID int64) (*passport.User, error) {
	user := &passport.User{}
	var err error
	cacheKey := u.GetCacheKeyUID(userID)
	fields := core.GetTableFilterFieldsFromCtx(ctx, (&passport.User{}).TableName())
	if fields == "*" {
		err = u.Redis.HGetAll(ctx, cacheKey).Scan(user)
	} else {
		err = u.Redis.HMGet(ctx, cacheKey, strings.Split(fields, ",")...).Scan(user)
	}

	// 将user 写入redis
	if err != nil {
		return nil, errors.WithCode(code.ErrRedis, err.Error())
	}

	if (passport.User{}) == *user {
		// key 不存在 读取db
		_ctx := core.SetTableFilterFieldsToCtx(ctx, user.TableName(), "*")
		user, err = u.UserRepo.GetUserByID(_ctx, userID)
		if err != nil {
			return user, err
		}

		if user != nil {
			_, err = u.Redis.HMSet(ctx, cacheKey, structs.Map(user)).Result()
			if err != nil {
				return nil, errors.WithCode(code.ErrRedis, err.Error())
			}
			_, err = u.Redis.Expire(ctx, cacheKey, UserTTL).Result()
			if err != nil {
				return user, errors.WithCode(code.ErrRedis, err.Error())
			}
		}
	}

	return user, nil
}

func CheckBUserLoginFromApp(ctx context.Context, accessToken string) (*passport.User, error) {
	// todo
	return nil, nil
}

func CheckBUserLoginFromPC(ctx context.Context, accessToken string) (*passport.User, error) {
	// todo
	return nil, nil
}

func (u userCache) CheckBUserAccessToken(ctx context.Context, accessToken string) (*passport.User, error) {

	curTime := time.Now().Unix()

	// 统一处理第一层缓存 token缓存
	id, err := u.accessToken2IDFormCache(ctx, accessToken)
	if err != nil {
		if errors.IsCode(err, code.ErrUserNotFound) {
			// 如果缓存里没有找到数据则读取数据库里的
			_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "*")
			user, err := u.UserRepo.GetUserByAccessToken(_ctx, accessToken)
			if err != nil {

				if errors.IsCode(err, code.ErrUserNotFound) {
					return user, errors.WithCode(code.ErrTokenInvalid, "无效的token")
				} else {
					return user, err
				}
			}

			// 检查db里存的token是否有效，目前为了兼容db处理，后续不存db后无需检查此项
			if user.ExpireTime < curTime {
				return user, errors.WithCode(code.ErrTokenInvalid, "无效的token")
			}

			// 设置token缓存
			err = u.AccessToken2IDToCache(ctx, accessToken, user.ID, user.ExpireTime-curTime+60)
			if err != nil {
				return user, err
			}

			id = user.ID
		} else {
			return nil, err
		}
	} else {
		// 有token缓存的情况
		// 兼容之前逻辑 距离过期时间不到30分钟时，则延长有效时间到一个小时
		cacheKey := u.GetCacheKeyToken(accessToken)
		ttl, err := u.Redis.TTL(ctx, cacheKey).Result()

		if err != nil {
			return nil, errors.WithCode(code.ErrRedis, err.Error())
		}

		if int64(ttl.Seconds()) < 1800 && int64(ttl) > 0 {
			// 先刷新uid对应信息里的expire_time，即使失败了下次也能继续走此处
			_, err := u.Redis.HSet(ctx, u.GetCacheKeyUID(id), "expire_time", curTime+3600).Result()
			if err != nil {
				return nil, errors.WithCode(code.ErrRedis, err.Error())
			}

			// 刷新token的有效期
			_, err = u.Redis.Expire(ctx, cacheKey, 3600*time.Second).Result()
			if err != nil {
				return nil, errors.WithCode(code.ErrRedis, err.Error())
			}
		}
	}

	// 统一获取一遍第二层的缓存 user_id为key获取数据
	user, err := u.GetUserByID(ctx, id)

	if err != nil {
		return user, err
	}

	// 此处统一记录当日登录用户缓存 todo 可以使用定时任务每天设置一次设置有效期,并且批量做一些延时刷到数据库的操作，最后登录时间和OS等 注意记得清除，别把内存撑爆了
	_cacheKey := u.GetCacheKeyLoginUsersByDate(time.Now().Format("2006-01-02"))
	_v, err := u.Redis.SIsMember(ctx, _cacheKey, user.ID).Result()
	if err != nil {
		return user, errors.WithCode(code.ErrRedis, err.Error())
	}
	if !_v {
		_, err = u.Redis.SAdd(ctx, _cacheKey, user.ID, user.ID).Result()
		if err != nil {
			return user, errors.WithCode(code.ErrRedis, err.Error())
		}

		_, err = u.Redis.Expire(ctx, _cacheKey, UserLoginListTTL).Result()
		if err != nil {
			return user, errors.WithCode(code.ErrRedis, err.Error())
		}
	}

	return user, err
}

func (u userCache) CheckMisUserAccessToken(ctx context.Context, accessToken string) (*passport.User, error) {
	user := &passport.User{}

	// 统一处理第一层缓存 token缓存
	cacheKey := u.GetCacheKeyMisToken(accessToken)
	val, err := u.Redis.Get(ctx, cacheKey).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.WithCode(code.ErrUserNotFound, err.Error(), cacheKey)
		}
		return nil, errors.WithCode(code.ErrRedis, err.Error(), cacheKey)
	}
	id, _ := strconv.ParseInt(val, 10, 64)

	// 统一获取一遍第二层的缓存 user_id为key获取数据
	user, err = u.GetUserByID(ctx, id)
	if err != nil {
		return user, err
	}

	return user, err
}

func (u userCache) accessToken2IDFormCache(ctx context.Context, accessToken string) (int64, error) {
	cacheKey := u.GetCacheKeyToken(accessToken)
	val, err := u.Redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, errors.WithCode(code.ErrUserNotFound, err.Error())
		}
		return 0, errors.WithCode(code.ErrRedis, err.Error())
	}
	id, _ := strconv.ParseInt(val, 10, 64)
	return id, nil
}

func (u userCache) AccessToken2IDToCache(ctx context.Context, accessToken string, userID int64, expireAfter int64) error {
	cacheKey := u.GetCacheKeyToken(accessToken)
	_, err := u.Redis.Set(ctx, cacheKey, userID, time.Duration(expireAfter)*time.Second).Result()

	if err != nil {
		return errors.WithCode(code.ErrRedis, err.Error())
	}
	return nil
}

/*
func (u userRepo) getUserByIDWithCacheKV(ctx context.Context, userID int64) (*domain.User, error) {
	user := &domain.User{}

	cacheKey := GenCacheKeyUID(userID)
	val, err := u.Redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// key 不存在 读取db
			ctx = context.WithValue(ctx, core.CtxKey("__fields"), domain.GetAllFieldsTblUser)
			user, err := u.GetUserByID(ctx, userID)
			if err != nil {
				return user, err
			}

			// 将user 写入redis
			uBytes, err := json.Marshal(user)
			if err != nil {
				return nil, errors.WithCode(code.ErrEncodingJSON, err.Error())
			}

			_, err = u.Redis.Set(ctx, cacheKey, string(uBytes), CacheTTL).Result()
			if err != nil {
				return nil, errors.WithCode(code.ErrRedis, err.Error())
			}

			return user, nil
		}

		return nil, errors.WithCode(code.ErrRedis, err.Error())
	}

	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		return nil, errors.WithCode(code.ErrDecodingJSON, err.Error())
	}

	return user, nil
}
*/

func (u userCache) DelAccessToken2IDFromCache(ctx context.Context, accessToken string) error {
	cacheKey := u.GetCacheKeyToken(accessToken)
	_, err := u.Redis.Del(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.WithCode(code.ErrUserNotFound, err.Error())
		}
		return errors.WithCode(code.ErrRedis, err.Error())
	}
	return nil
}

func (u userCache) MisAccessToken2IDToCache(ctx context.Context, accessToken string, userID int64, expireAfter int64) error {
	cacheKey := u.GetCacheKeyMisToken(accessToken)
	_, err := u.Redis.Set(ctx, cacheKey, userID, time.Duration(expireAfter)*time.Second).Result()

	if err != nil {
		return errors.WithCode(code.ErrRedis, err.Error())
	}
	return nil
}

func (u userCache) DelMisAccessToken2IDFromCache(ctx context.Context, accessToken string) error {
	cacheKey := u.GetCacheKeyMisToken(accessToken)
	_, err := u.Redis.Del(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.WithCode(code.ErrUserNotFound, err.Error())
		}
		return errors.WithCode(code.ErrRedis, err.Error())
	}
	return nil
}

func (u userCache) SetMisAccessTokenByUserID(ctx context.Context, accessToken string, userID int64, expireAfter int64) error {
	cacheKey := u.GetCacheKeyMisUserSession(userID)
	_, err := u.Redis.Set(ctx, cacheKey, accessToken, time.Duration(expireAfter)*time.Second).Result()

	if err != nil {
		return errors.WithCode(code.ErrRedis, err.Error())
	}
	return nil
}

func (u userCache) GetMisAccessTokenByUserID(ctx context.Context, userID int64) (string, error) {
	cacheKey := u.GetCacheKeyMisUserSession(userID)
	val, err := u.Redis.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errors.WithCode(code.ErrUserNotFound, err.Error())
		}
		return "", errors.WithCode(code.ErrRedis, err.Error())
	}
	return val, err
}

func (u userCache) CheckCUserAccessToken(ctx context.Context, accessToken string) (*passport.User, error) {

	curTime := time.Now().Unix()

	// 统一处理第一层缓存 token缓存
	id, err := u.accessToken2IDFormCache(ctx, accessToken)
	if err != nil {
		if errors.IsCode(err, code.ErrUserNotFound) {
			// 如果缓存里没有找到数据则读取数据库里的
			_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "*")
			user, err := u.UserRepo.GetUserByAccessToken(_ctx, accessToken)
			if err != nil {

				if errors.IsCode(err, code.ErrUserNotFound) {
					return user, errors.WithCode(code.ErrTokenInvalid, "无效的token")
				} else {
					return user, err
				}
			}

			// 检查db里存的token是否有效，目前为了兼容db处理
			if user.ExpireTime < curTime {
				return user, errors.WithCode(code.ErrTokenInvalid, "无效的token")
			}

			// 设置token缓存
			err = u.AccessToken2IDToCache(ctx, accessToken, user.ID, user.ExpireTime-curTime+60)
			if err != nil {
				return user, err
			}

			id = user.ID
		} else {
			return nil, err
		}
	}

	// 统一获取一遍第二层的缓存 user_id为key获取数据
	user, err := u.GetUserByID(ctx, id)

	if err != nil {
		return user, err
	}

	// 此处统一记录当日登录用户缓存 todo 可以使用定时任务每天设置一次设置有效期,并且批量做一些延时刷到数据库的操作，最后登录时间和OS等 注意记得清除，别把内存撑爆了
	_cacheKey := u.GetCacheKeyLoginUsersByDate(time.Now().Format("2006-01-02"))
	_v, err := u.Redis.SIsMember(ctx, _cacheKey, user.ID).Result()
	if err != nil {
		return user, errors.WithCode(code.ErrRedis, err.Error())
	}
	if !_v {
		_, err = u.Redis.SAdd(ctx, _cacheKey, user.ID, user.ID).Result()
		if err != nil {
			return user, errors.WithCode(code.ErrRedis, err.Error())
		}

		_, err = u.Redis.Expire(ctx, _cacheKey, UserLoginListTTL).Result()
		if err != nil {
			return user, errors.WithCode(code.ErrRedis, err.Error())
		}
	}

	return user, err
}

func (u userCache) DelUserCacheByUserID(ctx context.Context, userID int64) error {
	cacheKey := u.GetCacheKeyUID(userID)
	_, err := u.Redis.Del(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.WithCode(code.ErrUserNotFound, err.Error())
		}
		return errors.WithCode(code.ErrRedis, err.Error())
	}
	return nil
}
