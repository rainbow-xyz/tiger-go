package v1

import (
	"saas_service/internal/pkg/domain/passport"
	"saas_service/pkg/core"
)

type UserGetReq struct {
	core.MetaReqParams
	passport.User
	UserID int64 `json:"user_id" form:"user_id"`
}

type UserGetByTokenResp struct {
	ID         int64       `json:"id"`
	Phone      string      `json:"phone"`
	Name       string      `json:"name"`
	Sex        core.BitInt `json:"sex"`
	Unionid    string      `json:"unionid"`
	Nickname   string      `json:"nickname"`
	Headimgurl string      `json:"headimgurl"`
	Birthday   int64       `json:"birthday"`
}

// NewUserGetByTokenRespFromDomainUser 此处外部对象和内部领域对象的转换
func NewUserGetByTokenRespFromDomainUser(user *passport.User) *UserGetByTokenResp {

	if user != nil {
		return &UserGetByTokenResp{
			ID:         user.ID,
			Phone:      user.Phone,
			Name:       user.Name,
			Sex:        user.Sex,
			Unionid:    user.Unionid,
			Nickname:   user.Nickname,
			Headimgurl: user.Headimgurl,
			Birthday:   user.Birthday,
		}
	}
	return nil
}

type UserGetByIDResp struct {
	ID         int64  `json:"id"`
	Phone      string `json:"phone"`
	Name       string `json:"name"`
	Headimgurl string `json:"headimgurl"`
}

// NewUserGetByIDRespFromDomainUser 此处外部对象和内部领域对象的转换
func NewUserGetByIDRespFromDomainUser(user *passport.User) *UserGetByIDResp {

	if user != nil {
		return &UserGetByIDResp{
			ID:         user.ID,
			Phone:      user.Phone,
			Name:       user.Name,
			Headimgurl: user.Headimgurl,
		}
	}
	return nil
}

type MisLoginReq struct {
	core.MetaReqParams
	Phone string `json:"phone" form:"phone" uri:"phone"`
	Pwd   string `json:"pwd" form:"pwd" uri:"pwd"`
	Keep  int8   `json:"keep" form:"int8" uri:"int8"`
}

type misLoginResp struct {
	AccessToken string `json:"access_token"`
	Name        string `json:"name"`
	RealName    string `json:"real_name"`
	AvatarUrl   string `json:"avatar_url"`
	ExpireTime  int64  `json:"expire_time"`
}

func NewMisLoginResp(user *passport.User) *misLoginResp {
	if user != nil {
		return &misLoginResp{
			AccessToken: user.AccessToken,
			Name:        user.Name,
			RealName:    user.RealName,
			AvatarUrl:   user.AvatarUrl,
			ExpireTime:  user.ExpireTime,
		}
	}
	return nil
}

type LoginWxH5Req struct {
	core.MetaReqParams
	UnionID      string `json:"unionid" form:"unionid" uri:"unionid" validate:"required"`
	Nickname     string `json:"nickname" form:"nickname" uri:"nickname"`
	HeadImageUrl string `json:"headimgurl" form:"headimgurl" uri:"headimgurl"`
}

type loginWxH5Resp struct {
	ID           int64       `json:"id"`
	Phone        string      `json:"phone"`
	AccessToken  string      `json:"access_token"`
	Name         string      `json:"name"`
	Sex          core.BitInt `json:"sex"`
	RealName     string      `json:"real_name"`
	AvatarUrl    string      `json:"avatar_url"`
	ExpireTime   int64       `json:"expire_time"`
	HeadImageUrl string      `json:"headimgurl"`
}

func NewLoginWxH5Resp(user *passport.User) *loginWxH5Resp {
	if user != nil {
		return &loginWxH5Resp{
			ID:           user.ID,
			Phone:        user.Phone,
			AccessToken:  user.AccessToken,
			Name:         user.Name,
			Sex:          user.Sex,
			RealName:     user.RealName,
			AvatarUrl:    user.AvatarUrl,
			ExpireTime:   user.ExpireTime,
			HeadImageUrl: user.Headimgurl,
		}
	}
	return nil
}

type GetUserByUnionIDReq struct {
	core.MetaReqParams
	UnionID string `json:"unionid" form:"unionid" uri:"unionid"`
}

type getUserByUnionIDResp struct {
	ID    int64  `json:"id"`
	Phone string `json:"phone"`
}

func NewGetUserByUnionIDResp(user *passport.User) *getUserByUnionIDResp {
	if user != nil {
		return &getUserByUnionIDResp{
			ID:    user.ID,
			Phone: user.Phone,
		}
	}
	return nil
}
