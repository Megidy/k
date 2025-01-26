package user

import (
	"database/sql"
	"math/rand"

	"github.com/Megidy/k/types"
)

var pictures = []string{
	"https://i.pinimg.com/736x/b0/93/c5/b093c578f3c99b2525194db73cf12e01.jpg",
	"https://i.pinimg.com/736x/b2/7f/31/b27f31e8ddc96d54536e1f162948272a.jpg",
	"https://i.pinimg.com/736x/f8/b2/20/f8b220ee6d7f3c12b9c1ba3f202a5813.jpg",
	"https://i.pinimg.com/736x/6e/89/fa/6e89faaffc9ff42df7167c60abf6775c.jpg",
	"https://i.pinimg.com/736x/f0/07/f5/f007f57c6092bf3ca8189756de467365.jpg",
}

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
		err = row.Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.Description, &user.ProfilePicture)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}
func (s *store) CreateUser(user *types.User) error {
	user.ProfilePicture = pictures[rand.Intn(5)]

	_, err := s.db.Exec("insert into users(id,username,email,password,profile_picture) values(?,?,?,?,?)", user.ID, user.UserName, user.Email, user.Password, user.ProfilePicture)
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
		err = rows.Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.Description, &user.ProfilePicture)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}

func (s *store) UpdateDescription(userID, description string) error {
	_, err := s.db.Exec("update users set description=? where id=?", description, userID)

	if err != nil {
		return err
	}

	return nil
}
