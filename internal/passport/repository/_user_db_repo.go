package repository

import (
	"context"
	"database/sql"
	"saas_service/internal/pkg/domain/passport"
)

type userRepo struct {
	DB      *sql.DB
	Queries *Queries
}

const createUser = `-- name: CreateUser :one
INSERT INTO s_passport.user (
  id,
  name,
  phone,
  pwd,
  create_time,
  update_time
) VALUES (
  ?, ?, ?, ?, ?, ?
)
`

const GetUserBy = `-- name: CreateUser :one
Select ky_passport.user (
  id,
  name,
  phone,
  pwd,
  create_time,
  update_time
) VALUES (
  ?, ?, ?, ?, ?, ?
)
`

func NewUserRepo(db *sql.DB) passport.UserRepo {
	return &userRepo{
		DB:      db,
		Queries: NewQueries(db),
	}
}

func (u userRepo) CreateUser(ctx context.Context, user passport.User) (int64, error) {

	_, err := u.Queries.db.ExecContext(ctx, createUser,
		user.ID,
		user.Name,
		user.Phone,
		user.Pwd,
		user.CreateTime,
		user.UpdateTime,
	)

	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (u userRepo) GetUserByID(ctx context.Context, user passport.User) (int64, error) {

	_, err := u.Queries.db.ExecContext(ctx, createUser,
		user.ID,
		user.Name,
		user.Phone,
		user.Pwd,
		user.CreateTime,
		user.UpdateTime,
	)

	if err != nil {
		return 0, err
	}

	return user.ID, nil
}
