package usecase

import (
	"context"
	"github.com/marmotedu/errors"
	"saas_service/internal/pkg/code"
	"saas_service/internal/pkg/domain/passport"
	"saas_service/pkg/core"
	"time"
)

type brandUsecase struct {
	brandRepo         passport.BrandRepo
	brandExpireRepo   passport.BrandExpireRepo
	brandUniminiRepo  passport.BrandUniminiRepo
	brandMpExpireRepo passport.BrandMpExpireRepo
	brandAppRepo      passport.BrandAppRepo
}

func NewBrandUsecase(br passport.BrandRepo, ber passport.BrandExpireRepo, bum passport.BrandUniminiRepo, bmer passport.BrandMpExpireRepo, bar passport.BrandAppRepo) passport.BrandUsecase {
	return &brandUsecase{
		brandRepo:         br,
		brandExpireRepo:   ber,
		brandUniminiRepo:  bum,
		brandMpExpireRepo: bmer,
		brandAppRepo:      bar,
	}
}

func (u *brandUsecase) GetBrandByBrandID(ctx context.Context, brandID int64) (*passport.Brand, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Brand{}).TableName(), "id,name,db_name,create_user_id,create_time,update_time,logo")
	brand, err := u.brandRepo.GetBrandByBrandID(_ctx, brandID)
	return brand, err
}

func (u *brandUsecase) GetBrandUsersByUserID(ctx context.Context, userID int64) ([]passport.BrandUser, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandUser{}).TableName(), "id, brand_id, user_id")
	brandUsers, err := u.brandRepo.GetBrandUsersByUserID(_ctx, userID)
	return brandUsers, err
}

func (u *brandUsecase) BatchGetBrandsByBrandIDs(ctx context.Context, brandIDs []int64) ([]passport.Brand, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Brand{}).TableName(), "id,name,db_name,create_user_id,create_time,update_time,logo")
	brand, err := u.brandRepo.BatchGetBrandsByBrandIDs(_ctx, brandIDs)
	return brand, err
}

func (u *brandUsecase) GetBrandExpireByBrandID(ctx context.Context, brandID int64) (*passport.BrandExpire, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandExpire{}).TableName(), "id,brand_id,status,expire_time,level,last_sign_time")
	brandExpire, err := u.brandExpireRepo.GetBrandExpireByBrandID(_ctx, brandID)
	return brandExpire, err
}

func (u *brandUsecase) BatchGetBrandExpiresByBrandIDs(ctx context.Context, brandIDs []int64) ([]passport.BrandExpire, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandExpire{}).TableName(), "id,brand_id,status,expire_time,level,last_sign_time")
	brandExpires, err := u.brandExpireRepo.BatchGetBrandExpiresByBrandIDs(_ctx, brandIDs)
	return brandExpires, err
}

func (u *brandUsecase) GetBrandBaseAndLeaseInfoByUserID(ctx context.Context, userID int64) ([]passport.BrandBrandExpire, error) {
	brandInfo := make([]passport.BrandBrandExpire, 0)
	brandExpireMap := make(map[int64]passport.BrandExpire, 1)
	brandUsers, err := u.GetBrandUsersByUserID(ctx, userID)
	if err != nil || len(brandUsers) == 0 {
		return brandInfo, nil
	}

	var brandIDs []int64
	for _, v := range brandUsers {
		brandIDs = append(brandIDs, v.BrandID)
	}

	brands, err := u.BatchGetBrandsByBrandIDs(ctx, brandIDs)
	if err != nil || len(brands) == 0 {
		return brandInfo, nil
	}

	brandExpire, err := u.BatchGetBrandExpiresByBrandIDs(ctx, brandIDs)
	if err != nil {
		return brandInfo, nil
	}

	for _, v := range brandExpire {
		brandExpireMap[v.BrandID] = v
	}

	for _, v := range brands {
		var _be *passport.BrandExpire = nil
		_v, ok := brandExpireMap[v.ID]
		if ok {
			_be = &_v
		}
		brandInfo = append(brandInfo, passport.BrandBrandExpire{
			Brand:       v,
			BrandExpire: _be,
		})
	}

	return brandInfo, nil
}

func (u *brandUsecase) GetCreateValidBrandsByUserID(ctx context.Context, userID int64) ([]passport.Brand, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Brand{}).TableName(), "id,name,db_name,status,audit_status,audit_time,audit_reason")
	brands, err := u.brandRepo.GetCreateValidBrandsByUserID(_ctx, userID)
	return brands, err
}

func (u *brandUsecase) GetAllNormalExpireBrandIDs(ctx context.Context) ([]string, error) {
	brandIDs, err := u.brandExpireRepo.GetAllNormalExpireBrandIDs(ctx)
	return brandIDs, err
}

func (u *brandUsecase) CheckCUserBrandRelation(ctx context.Context, brandID int64, userID int64) (bool, error) {
	return u.brandRepo.CheckBrandUsersByBrandIDAndUserID(ctx, brandID, userID)
}

func (u *brandUsecase) AddBrandUserRelation(ctx context.Context, brandID int64, userID int64) (int64, error) {
	return u.brandRepo.AddBrandUserRelation(ctx, brandID, userID)
}

func (u *brandUsecase) GetUniminiExpireByBrandID(ctx context.Context, brandID int64) (*passport.BrandUnimini, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandUnimini{}).TableName(), "expire_time")
	return u.brandUniminiRepo.GetUniminiExpireByBrandID(_ctx, brandID)
}

func (u *brandUsecase) CheckUniminiExpire(ctx context.Context, brandID int64) error {
	curTime := time.Now().Unix()

	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandExpire{}).TableName(), "expire_time")
	brandExpires, err := u.brandExpireRepo.GetBrandExpireByBrandID(_ctx, brandID)
	if err != nil {
		if errors.IsCode(err, code.ErrDataNotFound) {
			err = errors.WithCode(code.ErrBrandNotFound, "品牌未找到")
			return err
		}
		return err
	}

	if brandExpires.ExpireTime < curTime {
		err = errors.WithCode(code.ErrBrandNotFound, "品牌未找到")
		return err
	}

	_ctx = core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandUnimini{}).TableName(), "expire_time")
	bum, err := u.brandUniminiRepo.GetUniminiExpireByBrandID(_ctx, brandID)
	if err != nil {
		if errors.IsCode(err, code.ErrDataNotFound) {
			err = errors.WithCode(code.ErrUniminiNotOpen, "未开通通用小程序")
			return err
		}
		return err
	}

	if bum.ExpireTime < curTime {
		err = errors.WithCode(code.ErrUniminiExpired, "通用小程序已过期")
		return err
	}
	return nil
}

func (u *brandUsecase) GetBrandFullInfoByBrandID(ctx context.Context, brandID int64) (*passport.Brand, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Brand{}).TableName(), "*")
	brand, err := u.brandRepo.GetBrandByBrandID(_ctx, brandID)
	return brand, err
}

func (u *brandUsecase) GetBrandMpExpireByBrandID(ctx context.Context, brandID int64) (*passport.BrandMpExpire, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandMpExpire{}).TableName(), "id,brand_id,status,expire_time,level,last_sign_time")
	obj, err := u.brandMpExpireRepo.GetBrandMpExpireByBrandID(_ctx, brandID)
	return obj, err
}

func (u *brandUsecase) GetBrandAppByBrandID(ctx context.Context, brandID int64) (*passport.BrandApp, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.BrandApp{}).TableName(),
		"id, appid, pay_authed ,authed")
	obj, err := u.brandAppRepo.GetBrandAppByBrandID(_ctx, brandID)

	return obj, err
}
