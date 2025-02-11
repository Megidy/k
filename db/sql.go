package db

import (
	"database/sql"

	"github.com/Megidy/k/config"
	_ "github.com/go-sql-driver/mysql"
)

func NewSQlDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.MySQLConnectionString)
	if err != nil {
		panic(err)
	}
	err = pingDB(db)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil

}

func pingDB(db *sql.DB) error {
	return db.Ping()
}
