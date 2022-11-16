package repository

import (
	"context"
	"saas_service/internal/pkg/code"
	"saas_service/internal/pkg/constants"
	"saas_service/internal/pkg/domain/passport"
	"saas_service/pkg/core"

	"github.com/marmotedu/errors"
	"gorm.io/gorm"
)

type userRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) passport.UserRepo {
	return &userRepo{
		DB: db,
	}
}

// LocalRepoAtomic 本repository层的事务使用，供usecase层调用，这属于代码分层db事务的方式二，实现起来略微复杂，引入回调（跨repository的事务如何设计还在研究思考中）
func (u userRepo) LocalRepoAtomic(ctx context.Context, fn func(u passport.UserRepo) (interface{}, error)) (result interface{}, err error) {
	trx := u.DB.Begin()

	// 处理异常情况
	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			err = errors.WithCode(code.ErrDatabase, "Exec database transaction failed")
			panic(p)
		}

		if err != nil {
			trx.Rollback()
		} else {
			trx.Commit()
		}
	}()

	// log.Println("....LocalRepoAtomic 1 ...", &u, u)
	u.DB = trx // copy一个新的userRepo给外面的实际执行事务过程的函数用
	// log.Println("....LocalRepoAtomic 2 ...", &u, u)
	result, err = fn(u)
	return
}

// WithTrx 返回一个持有事务句柄的repository的拷贝，这属于代码分层实现db事务的方式一，但是对上次暴露了更多细节 使用要克制
func (u userRepo) WithTrx(trx *gorm.DB) passport.UserRepo {
	if trx == nil {
		return u
	}

	// 此处接收一个外部传入的trx，并且返回个u的拷贝
	u.DB = trx
	return u
}

func (u userRepo) GetTrx() *gorm.DB {
	return u.DB.Begin()
}

func (u userRepo) GetUserByID(ctx context.Context, userID int64) (*passport.User, error) {
	user := &passport.User{}
	err := u.DB.WithContext(ctx).Table(user.TableName()).
		Select(core.GetTableFilterFieldsFromCtx(ctx, user.TableName())).
		Where("id = ? and status = ?", userID, constants.UserStatusOk).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrUserNotFound, err.Error())
		}

		return nil, errors.WithCode(code.ErrDatabase, err.Error())
	}

	return user, nil
}

func (u userRepo) GetUserByPhone(ctx context.Context, phone string) (*passport.User, error) {
	cond := []interface{}{" phone = ? AND status != ?", phone, constants.UserStatusDel}
	return u.getSingleUserByCond(ctx, cond)
}

func (u userRepo) getSingleUserByCond(ctx context.Context, cond []interface{}) (*passport.User, error) {
	// 为防止出现条件遗漏 禁止传入空条件 此表禁止全量查询
	if len(cond) == 0 {
		cond = append(cond, "1 = -1")
	}

	queryPreParam := cond[0]
	queryParamVal := cond[1:]

	user := &passport.User{}
	err := u.DB.WithContext(ctx).Table(user.TableName()).
		Select(core.GetTableFilterFieldsFromCtx(ctx, user.TableName())).
		Where(queryPreParam, queryParamVal...).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrUserNotFound, err.Error())
		}

		return nil, errors.WithCode(code.ErrDatabase, err.Error())
	}

	return user, nil
}

// UpdateUserByCondUseMap 根据条件更新user信息  tips：设计考量 conds尽可能的显式构造条件，安全可靠，便于维护，并且给予一定的灵活性，兼容掌馆系统的习惯，
// 实体对象数据没有用结构体而是使用了field和val的map，避免gorm对零值的一些特殊处理，兼容php的习惯，后期在实践中逐步改进
func (u userRepo) UpdateUserByCondUseMap(ctx context.Context, entity map[string]interface{}, cond []interface{}) (int64, error) {
	// 为防止出现条件遗漏 禁止传入空条件 禁止更新全表数据
	if len(cond) == 0 {
		cond = append(cond, "1 = -1")
	}

	queryPreParam := cond[0]
	queryParamVal := cond[1:]

	// 示例1：只输出sql不执行 调试使用
	/*
		sql := u.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Table((&passport.User{}).TableName()).Where(queryPreParam, queryParamVal...).Updates(entity)
		})
	*/

	// 示例2：执行sql并写入日志 调试使用
	// result := u.DB.Debug().WithContext(ctx).Table((&passport.User{}).TableName()).Where(queryPreParam, queryParamVal...).Updates(entity)

	result := u.DB.WithContext(ctx).Table((&passport.User{}).TableName()).Where(queryPreParam, queryParamVal...).Updates(entity)
	err := result.Error
	if result.Error != nil {
		return result.RowsAffected, errors.WithCode(code.ErrDatabase, err.Error())
	}

	return result.RowsAffected, nil

}

// UpdateUserByCond 根据条件更新user信息  tips：设计考量 conds尽可能的显式构造条件，安全可靠，便于维护，并且给予一定的灵活性，兼容掌馆系统的习惯，
// 实体对象数据更换为结构体实现，同时为了避免gorm对零值的处理，以及保护某些字段被误改，所以要显式指定要改哪些字段
func (u userRepo) UpdateUserByCond(ctx context.Context, user *passport.User, cond []interface{}) (int64, error) {
	// 为防止出现条件遗漏 禁止传入空条件 禁止更新全表数据
	if len(cond) == 0 {
		cond = append(cond, "1 = -1")
	}

	queryPreParam := cond[0]
	queryParamVal := cond[1:]

	// 示例1：只输出sql不执行 调试使用
	/*
		sql := u.DB.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Table((&passport.User{}).TableName()).Where(queryPreParam, queryParamVal...).Updates(user)
		})
		log.Println(sql)
	*/

	filter := core.GetTableCUFilterFieldsFromCtx(ctx, user.TableName())
	result := u.DB.Debug().WithContext(ctx).Table((&passport.User{}).TableName()).
		Select(filter[0], filter[1:]).
		Where(queryPreParam, queryParamVal...).Updates(user)

	err := result.Error
	if result.Error != nil {
		return result.RowsAffected, errors.WithCode(code.ErrDatabase, err.Error())
	}

	return result.RowsAffected, nil

}

func (u userRepo) UpdateUserByIDUseMap(ctx context.Context, entity map[string]interface{}, id int64) (int64, error) {
	cond := []interface{}{" id = ?", id}
	return u.UpdateUserByCondUseMap(ctx, entity, cond)

}
func (u userRepo) UpdateUserByID(ctx context.Context, user *passport.User, userID int64) (rowsAffected int64, err error) {
	cond := []interface{}{" id = ?", userID}
	rowsAffected, err = u.UpdateUserByCond(ctx, user, cond)
	return
}

func (u userRepo) AddUser(ctx context.Context, user *passport.User) (int64, error) {
	filter := core.GetTableCUFilterFieldsFromCtx(ctx, user.TableName())
	result := u.DB.WithContext(ctx).
		Select(filter[0], filter[1:]).
		Table((&passport.User{}).TableName()).Create(user)
	err := result.Error
	if err != nil {
		return 0, errors.WithCode(code.ErrDatabase, err.Error())
	}

	return user.ID, err
}

func (u userRepo) GetUserByAccessToken(ctx context.Context, accessToken string) (*passport.User, error) {
	cond := []interface{}{" access_token = ? AND status = ? ", accessToken, constants.UserStatusOk}
	return u.getSingleUserByCond(ctx, cond)
}
