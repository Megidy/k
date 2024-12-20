package user

import (
	"database/sql"

	"github.com/Megidy/k/types"
)

type store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *store {
	return &store{db: db}
}

func (s *store) GetUserById(id string) (*types.User, error) {
	row, err := s.db.Query("select * from users where id=?", id)
	if err != nil {
		return nil, err
	}
	var user types.User
	for row.Next() {
		err = row.Scan(&user.ID, &user.UserName, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}
