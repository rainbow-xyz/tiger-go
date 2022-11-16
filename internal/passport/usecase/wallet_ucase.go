package usecase

import (
	"context"
	"gorm.io/gorm"
	"saas_service/internal/pkg/domain/passport"
	"saas_service/pkg/core"
)

type walletUsecase struct {
	walletRepo passport.WalletRepo
}

func NewWalletUsecase(sa passport.WalletRepo) passport.WalletUsecase {
	return &walletUsecase{
		walletRepo: sa,
	}
}

func (wa *walletUsecase) WalletGetByCond(ctx context.Context, fields string, cond string) ([]passport.Wallet, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Wallet{}).TableName(), fields)
	wallet, err := wa.walletRepo.WalletGetByCond(_ctx, cond)
	return wallet, err
}

func (wa *walletUsecase) WalletGetSingleByCond(ctx context.Context, fields string, cond string) (*passport.Wallet, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Wallet{}).TableName(), fields)
	wallet, err := wa.walletRepo.WalletGetSingleByCond(_ctx, cond)
	return wallet, err
}

func (wa *walletUsecase) WalletGetByID(ctx context.Context, id int) (*passport.Wallet, error) {
	_ctx := core.SetTableFilterFieldsToCtx(ctx, (&passport.Wallet{}).TableName(), "*")
	wallet, err := wa.walletRepo.WalletGetByID(_ctx, id)
	return wallet, err
}

func (wa *walletUsecase) WalletEditByIDEx(ctx context.Context, wallet passport.Wallet) (int64, error) {
	updateMap := make(map[string]interface{})
	if wallet.Balance != "" {
		updateMap["balance"] = gorm.Expr(wallet.Balance)
	}
	rows, err := wa.walletRepo.WalletEditByIDEx(ctx, wallet.ID, updateMap)
	return rows, err
}
