package code

// 错误码设计说明
// saas-service: passport_server errors.
const (
	// ErrUserNotFound - 404: User not found.
	ErrUserNotFound int = iota + 110001

	// ErrUserAlreadyExist - 500: User already exist.
	ErrUserAlreadyExist

	ErrUserNotRelateBrand

	// ErrWxUserIsNotRegistered  - 500： WeChat user is not registered
	ErrWxUserIsNotRegistered

	// ErrThePhoneHasBeenBoundToOtherWeChatUser - 500: the phone has been bound to other Wechat user
	ErrThePhoneHasBeenBoundToOtherWeChatUser

	// ErrBrandNotFound 400: Brand not found
	ErrBrandNotFound

	// ErrUniminiNotOpen 500: 小程序未开通
	ErrUniminiNotOpen

	// ErrBrandExpired 500：品牌已过期
	ErrBrandExpired

	// ErrUniminiExpired 500：小程序已过期
	ErrUniminiExpired

	// ErrMiniProgramUnauthorized 500：定制小程序未授权给三方平台
	ErrMiniProgramUnauthorized

	// ErrMiniProgramNotOpenPay 500：定制小程序未开通支付
	ErrMiniProgramNotOpenPay
)
