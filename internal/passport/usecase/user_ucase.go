package usecase

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/marmotedu/errors"
	"gorm.io/gorm"
	"saas_service/internal/pkg/code"
	constants "saas_service/internal/pkg/constants"
	"saas_service/internal/pkg/domain/passport"
	corecs "saas_service/pkg/constants"
	"saas_service/pkg/core"
	"saas_service/pkg/xlog"
	"strings"
	"time"
)

const MisAccessTokenCacheTTL = 7 * 86400
const WxH5AccessTokenTTL = 30 * 86400

type userUsecase struct {
	userRepo    passport.UserRepo
	misUserRepo passport.MisUserRepo
	userCache   passport.UserCache
	agentRepo   passport.AgentRepo
	brandRepo   passport.BrandRepo
}

func NewUserUsecase(u passport.UserRepo, mu passport.MisUserRepo, uc passport.UserCache, a passport.AgentRepo, b passport.BrandRepo) passport.UserUsecase {
	return &userUsecase{
		userRepo:    u,
		misUserRepo: mu,
		userCache:   uc,
		agentRepo:   a,
		brandRepo:   b,
	}
}

func (u *userUsecase) GetUserByID(ctx context.Context, userID int64) (*passport.User, error) {
	user, err := u.userCache.GetUserByID(ctx, userID)
	return user, err
}

func (u *userUsecase) CheckBUserLogin(ctx context.Context, accessToken string) (*passport.User, error) {
	// 根据token查看用户相关信息
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "id,phone,name,sex,unionid,headimgurl,nickname,birthday")
	user, err := u.userCache.CheckBUserAccessToken(_ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return user, err
}

func (u *userUsecase) CheckMisUserLogin(ctx context.Context, accessToken string) (*passport.User, error) {
	// 根据token查看用户相关信息
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "id,phone,name,sex,unionid,headimgurl,nickname")
	user, err := u.userCache.CheckMisUserAccessToken(_ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if user.ID != 0 {
		// 检查是否是mis用户
		_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.MisUser{}).TableName(), "id,user_id,name,phone")
		misUser, err := u.misUserRepo.GetMisUserByUserID(_ctx, user.ID)
		if err != nil {
			return nil, err
		}

		if misUser == nil || misUser.UserID == 0 {
			return nil, errors.WithCode(code.ErrUserNotFound, "用户未找到")
		}
	}

	return user, err
}

func (u *userUsecase) BLogout(ctx context.Context, accessToken string) error {
	u.userCache.DelAccessToken2IDFromCache(ctx, accessToken)
	return nil
}

func (u *userUsecase) BSetLoginInfo(ctx context.Context, accessToken string) (*passport.User, error) {
	user, err := u.CheckBUserLogin(ctx, accessToken)
	if user != nil && user.ID != 0 {
		DoCacheOpWithRetry[int64](u.userCache.DelUserCacheByUserID, 3)(ctx, user.ID)
	}

	// todo 其它处理
	return user, err
}

func (u *userUsecase) MisLogin(ctx context.Context, phone string, pwd string, keep int8) (*passport.User, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "id,name,avatar_url,real_name")
	m := md5.New()
	m.Write([]byte(pwd))
	pwd = hex.EncodeToString(m.Sum(nil))
	user, err := u.userRepo.GetUserByPhoneAndPwd(_ctx, phone, pwd)
	if err != nil {
		if errors.IsCode(err, code.ErrUserNotFound) {
			return nil, errors.WithCode(code.ErrPasswordIncorrect, err.Error())
		}
		return nil, err
	}

	if user != nil && user.ID != 0 {
		// 检查是否是mis用户
		_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.MisUser{}).TableName(), "id,user_id,name,phone")
		misUser, err := u.misUserRepo.GetMisUserByUserID(_ctx, user.ID)
		if err != nil {
			return nil, err
		}

		if misUser == nil || misUser.UserID == 0 {
			return nil, errors.WithCode(code.ErrUserNotFound, "用户未找到")
		}
	}

	prevAccessToken, err := u.userCache.GetMisAccessTokenByUserID(ctx, user.ID)

	// 生成新的token
	accessToken := core.GenerateAccessToken()
	var timeAfter int64 = MisAccessTokenCacheTTL
	if keep > 0 {
		timeAfter += 30 * 86400
	}
	err = u.userCache.MisAccessToken2IDToCache(ctx, accessToken, user.ID, timeAfter)
	if err != nil {
		return nil, err
	}
	user.AccessToken = accessToken

	err = u.userCache.SetMisAccessTokenByUserID(ctx, accessToken, user.ID, timeAfter)
	if err != nil {
		return nil, err
	}

	if prevAccessToken != "" {
		err = u.userCache.DelMisAccessToken2IDFromCache(ctx, prevAccessToken)
		if err != nil {
			return nil, err
		}
	}
	user.ExpireTime = time.Now().Unix() + timeAfter

	return user, err
}

func (u *userUsecase) CheckBUserAgentLogin(ctx context.Context, accessToken string) (*passport.User, error) {
	// 根据token查看用户相关信息
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "id,phone,name,sex,unionid,headimgurl,nickname,birthday")
	user, err := u.userCache.CheckBUserAccessToken(_ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// 检查是否是agent管理员
	_ctx = core.SetTableFilterFieldsToCtx(ctx, (&passport.Agent{}).TableName(), "id")
	result, err := u.agentRepo.CheckAgentAdminByUserID(_ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if result == false {
		return nil, errors.WithCode(code.ErrPermissionDenied, "无权限登录")
	}

	return user, err
}

func (u *userUsecase) CheckCUserLogin(ctx context.Context, accessToken string, brandID int64) (*passport.User, error) {
	// 根据token查看用户相关信息
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "id,phone,name,sex,unionid,headimgurl,nickname,birthday")
	user, err := u.userCache.CheckCUserAccessToken(_ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// 检查用户品牌关系绑定
	if brandID != 0 {
		result, err := u.brandRepo.CheckBrandUsersByBrandIDAndUserID(ctx, brandID, user.ID)
		if err != nil {
			return nil, err
		}

		if result == false {
			return nil, errors.WithCode(code.ErrUserNotRelateBrand, "用户未关联该品牌")
		}
	}

	return user, err
}

// LoginWxH5 微信H5端登录
func (u *userUsecase) LoginWxH5(ctx context.Context, param *passport.User) (*passport.User, error) {

	// 查看用户是否存在
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
		"id,phone,name,sex,avatar_url,access_token,unionid,headimgurl,nickname,birthday,expire_time,login_times")
	user, err := u.userRepo.GetUserByUnionID(_ctx, param.Unionid)

	if err != nil {

		if errors.IsCode(err, code.ErrUserNotFound) {
			// 该微信用户未绑定掌馆用户
			err = errors.WithCode(code.ErrWxUserIsNotRegistered, "微信用户未注册")
		}
		return nil, err
	}

	curTime := time.Now().Unix()
	// 写回数据库并且刷新user信息缓存
	entity := map[string]interface{}{
		"last_login_time": curTime,
		"update_time":     curTime,
		"login_times":     gorm.Expr("login_times + ?", 1),
	}

	// 登录系统获取token
	if user.ExpireTime < curTime {
		oldToken := user.AccessToken

		user.AccessToken = core.GenerateAccessToken()
		user.ExpireTime = curTime + WxH5AccessTokenTTL

		entity["access_token"] = user.AccessToken
		entity["expire_time"] = user.ExpireTime

		// 清除原有token
		err := u.userCache.DelAccessToken2IDFromCache(ctx, oldToken)
		if err != nil {
			// todo 此处可以考虑添加重试机制，不然可能会导致数据不一致
		}
	}

	// 其它额外信息
	if param.Headimgurl != "" {
		user.Headimgurl = param.Headimgurl
		entity["headimgurl"] = param.Headimgurl

		// 刷新 未设置用户头像的 C端用户头像
		if user.AvatarUrl != "" &&
			!strings.Contains(user.AvatarUrl, corecs.OSSDomainOnline) &&
			!strings.Contains(user.AvatarUrl, corecs.OSSDomainOffline) {

			user.AvatarUrl = param.Headimgurl
			entity["avatar_url"] = param.Headimgurl
		}
	}

	if param.Nickname != "" {
		entity["nickname"] = param.Nickname
		user.Nickname = param.Nickname
	}

	_, err = u.userRepo.UpdateUserByIDUseMap(ctx, entity, user.ID)
	if err != nil {
		return nil, err
	}

	DoCacheOpWithRetry[int64](u.userCache.DelUserCacheByUserID, 3)(ctx, user.ID)

	return user, nil
}

func (u *userUsecase) GetUserByUnionID(ctx context.Context, unionID string) (*passport.User, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "id,phone")
	return u.userRepo.GetUserByUnionID(_ctx, unionID)
}

// RegisterCUserByWeChatUser 通过微信用户注册C端user 该接口演示了目前阶段思考的db事务的使用方法
func (u *userUsecase) RegisterCUserByWeChatUser(ctx context.Context, param *passport.User) (*passport.User, error) {
	// todo 此处缺失对tbl_user_wx 表的同步处理，因为目前并未有地方用到这个表，如有需要再补上
	/*
		eg: 事务示例代码方式一 目前为了代码分层上的统一，暴露出db对象来，后续继续研究考虑更合理的实现方式
		// trx := u.userRepo.GetTrx()
		// user, err := u.userRepo.WithTrx(trx).GetUserByUnionID(_ctx, param.Unionid)
		// trx.Commit()

		目前实际使用的是下列方式
	*/

	var delUserCacheByUserID int64 = 0

	// log.Println("RegisterCUserByWeChatUser....", u.userRepo)
	// 原子操作执行代码块

	atomicBlock := func(u passport.UserRepo) (interface{}, error) {
		// 根据微信unionid判断用户是否存在
		// log.Println("atomicBlock....", u)
		curTime := time.Now().Unix()
		_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
			"id")
		user, err := u.GetUserByUnionIDIgnoreStatusCheck(_ctx, param.Unionid)
		if err == nil {
			// 用户存在直接返回
			return user, nil
		} else if !errors.IsCode(err, code.ErrUserNotFound) {
			return nil, err
		}

		// 根据手机号判断user是否存在
		_ctx = core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
			"id,unionid")
		user, err = u.GetUserByPhoneIgnoreStatusCheck(_ctx, param.Phone)
		if err != nil {
			if !errors.IsCode(err, code.ErrUserNotFound) {
				return nil, err
			}

			// 如果用户数据不存在则添加
			entity := &passport.User{
				Unionid:    param.Unionid,
				Phone:      param.Phone,
				Nickname:   param.Nickname,
				Name:       param.Nickname,
				RealName:   param.Nickname,
				Headimgurl: param.Headimgurl,
				AvatarUrl:  param.Headimgurl,
				Sex:        param.Sex,
				CreateTime: curTime,
				UpdateTime: curTime,
			}
			// 此处必须指定要筛选的字段，否则可能导致某些字段使用了结构体的默认值 如果不好维护的话 可以换成map形式的简单可靠
			_ctx = core.SetTableCUFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
				"unionid,phone,nickname,name,real_name,headimgurl,avatar_url,sex,create_time,update_time")

			id, err := u.AddUser(_ctx, entity)
			if err != nil {
				return nil, err
			}

			return &passport.User{ID: id}, nil

		} else {
			// 用户数据已存在
			if user.Unionid != "" {
				// 该手机号已被其它微信账号绑定
				err = errors.WithCode(code.ErrThePhoneHasBeenBoundToOtherWeChatUser, "该手机号已被其它微信账号绑定")
				return nil, err
			}

			// 绑定unionid
			entity := &passport.User{
				Unionid:    param.Unionid,
				Nickname:   param.Nickname,
				Headimgurl: param.Headimgurl,
				UpdateTime: curTime,
			}
			// 此处需显式指定字段名称，否则无法达到预期结果（不传的话 编辑时不会更改任何内容，接口不会报错）
			_ctx = core.SetTableCUFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "unionid,nickname,headimgurl,update_time")
			cond := []interface{}{" id = ? AND (unionid = \"\" OR unionid IS NULL)", user.ID}
			_ctx := core.SetTableFilterFieldsToCtx(_ctx, (&passport.User{}).TableName(), "*")
			_, err := u.UpdateUserByCond(_ctx, entity, cond)
			if err != nil {
				return nil, err
			}
			delUserCacheByUserID = user.ID
			return &passport.User{ID: user.ID}, nil
		}
		return nil, nil
	}

	result, err := u.userRepo.LocalRepoAtomic(ctx, atomicBlock)

	if err != nil {
		return nil, err
	}

	// 清空redis里的缓存
	_, _err := DoCacheOpWithRetry[int64](u.userCache.DelUserCacheByUserID, 3)(ctx, delUserCacheByUserID)
	if _err != nil {
		xlog.XSErrorF(ctx, "清除用户(uid:%d)缓存信息失败", delUserCacheByUserID, _err)
	}
	return result.(*passport.User), err
}

func DoCacheOpWithRetry[T any](f func(ctx context.Context, cacheKey T) error, maxTryCount int) func(ctx context.Context, cacheKey T) (tryCount int, err error) {
	return func(ctx context.Context, cacheKey T) (tryCount int, err error) {

		for tryCount = 1; tryCount <= maxTryCount; tryCount++ {
			err = f(ctx, cacheKey)
			if err == nil {
				return tryCount, err
			}
		}

		// todo 完善重试机制

		return tryCount, err
	}
}

// BoundPhoneForWeChatCUser 为微信C端用户绑定手机号
func (u *userUsecase) BoundPhoneForWeChatCUser(ctx context.Context, param *passport.User) error {
	curTime := time.Now().Unix()

	// 检查该手机号用户是否已注册过user
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
		"id,unionid,phone")
	user, err := u.userRepo.GetUserByUnionIDIgnoreStatusCheck(_ctx, param.Unionid)
	if err != nil && !errors.IsCode(err, code.ErrUserNotFound) {
		return err
	}

	if user != nil && user.Unionid != "" && user.Unionid != param.Unionid {
		// 该手机号已被其它微信账号绑定
		err = errors.WithCode(code.ErrThePhoneHasBeenBoundToOtherWeChatUser, "该手机号已被其它微信账号绑定")
		return err
	}

	// 绑定phone 只能绑定没有手机号的  已有的不作处理直接返回成功
	entity := &passport.User{
		Phone:      param.Phone,
		UpdateTime: curTime,
	}
	cond := []interface{}{" id = ? AND (phone = \"\" OR phone IS NULL) AND status != ?", user.ID, constants.UserStatusDel}
	// 此处需显式指定字段名称，否则无法达到预期结果（不传的话 编辑时不会更改任何内容，接口不会报错）
	_ctx = core.SetTableCUFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "phone,update_time")
	_, err = u.userRepo.UpdateUserByCond(_ctx, entity, cond)
	if err != nil {
		return err
	}

	// 清空redis里的缓存
	_, _err := DoCacheOpWithRetry[int64](u.userCache.DelUserCacheByUserID, 3)(ctx, user.ID)
	if _err != nil {
		xlog.XSErrorF(ctx, "清除用户(uid:%d)缓存信息失败", user.ID, _err)
	}

	return nil
}

// BoundAdditionalInfoForWeChatCUser 补充微信端手机号等用户信息
func (u *userUsecase) BoundAdditionalInfoForWeChatCUser(ctx context.Context, param *passport.User) error {
	curTime := time.Now().Unix()

	// 检查该手机号用户是否已注册过user
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
		"id,unionid,phone")
	user, err := u.userRepo.GetUserByUnionIDIgnoreStatusCheck(_ctx, param.Unionid)
	if err != nil && !errors.IsCode(err, code.ErrUserNotFound) {
		return err
	}

	if user != nil && user.Unionid != "" && user.Unionid != param.Unionid {
		// 该手机号已被其它微信账号绑定
		err = errors.WithCode(code.ErrThePhoneHasBeenBoundToOtherWeChatUser, "该手机号已被其它微信账号绑定")
		return err
	}

	// 绑定phone 只能绑定没有手机号的  已有的不作处理直接返回成功
	entity := &passport.User{
		Phone:           param.Phone,
		Name:            param.Name,
		Sex:             param.Sex,
		Birthday:        param.Birthday,
		PersonSignature: param.PersonSignature,
		UpdateTime:      curTime,
	}
	_ctx = core.SetTableCUFilterFieldsToCtx(ctx, (&passport.User{}).TableName(), "phone,name,sex,birthday,person_signature,update_time")

	cond := []interface{}{" id = ? AND (phone = \"\" OR phone IS NULL) AND status != ?", user.ID, constants.UserStatusDel}
	// 此处需显式指定字段名称，否则无法达到预期结果（不传的话 编辑时不会更改任何内容，接口不会报错）
	_, err = u.userRepo.UpdateUserByCond(_ctx, entity, cond)
	if err != nil {
		return err
	}

	// 清空redis里的缓存
	_, _err := DoCacheOpWithRetry[int64](u.userCache.DelUserCacheByUserID, 3)(ctx, user.ID)
	if _err != nil {
		xlog.XSErrorF(ctx, "清除用户(uid:%d)缓存信息失败", user.ID, _err)
	}

	return nil
}

// UnbindWeChatCUser 微信C端用户解绑
func (u *userUsecase) UnbindWeChatCUser(ctx context.Context, accessToken string) error {
	curTime := time.Now().Unix()
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
		"id")
	user, err := u.userRepo.GetUserByAccessTokenIgnoreStatusCheck(_ctx, accessToken)
	if err != nil && !errors.IsCode(err, code.ErrUserNotFound) {
		return err
	}

	if user != nil && user.ID != 0 {
		// 清除unionid
		entity := map[string]interface{}{
			"unionid":     nil,
			"expire_time": curTime - 10,
			"update_time": curTime,
		}
		cond := []interface{}{" id = ?", user.ID}
		// 此处需显式指定字段名称，否则无法达到预期结果（不传的话 编辑时不会更改任何内容，接口不会报错）
		_, err = u.userRepo.UpdateUserByCondUseMap(_ctx, entity, cond)
		if err != nil {
			return err
		}

		// 清除缓存 如果失败了此处用户端也可以重试
		_, err = DoCacheOpWithRetry[int64](u.userCache.DelUserCacheByUserID, 3)(ctx, user.ID)
		if err != nil {
			return err
		}

		// 清除缓存 如果失败了此处用户端也可以重试
		_, err = DoCacheOpWithRetry[string](u.userCache.DelAccessToken2IDFromCache, 3)(ctx, accessToken)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *userUsecase) GetUserByAccessToken(ctx context.Context, accessToken string) (*passport.User, error) {
	// 根据token查看用户相关信息
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
		"*")
	user, err := u.userRepo.GetUserByAccessToken(_ctx, accessToken)
	return user, err
}

func (u *userUsecase) GetUserByPhone(ctx context.Context, phone string) (*passport.User, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.User{}).TableName(),
		"id")
	user, err := u.userRepo.GetUserByPhone(_ctx, phone)
	return user, err
}
