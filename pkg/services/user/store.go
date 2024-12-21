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
func (s *store) CreateUser(user *types.User) error {
	_, err := s.db.Exec("insert into users values(?,?,?,?)", user.ID, user.UserName, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}
func (s *store) UserExists(user *types.User) (bool, error) {
	rows, err := s.db.Query("select * from users where email=? or username=?", user.Email, user.UserName)
	if err != nil {
		return false, err
	}
	for !rows.Next() {
		return false, nil
	}
	return true, nil
}

func (s *store) GetUserByEmail(email string) (*types.User, error) {
	rows, err := s.db.Query("select * from users where email=?", email)
	if err != nil {
		return nil, err
	}
	var user types.User
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.UserName, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}
