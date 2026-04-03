package db

import (
	"database/sql"
	"fmt"

	"github.com/esuEdu/investment-risk-engine/pkg/logger"
	_ "github.com/lib/pq"
)

func NewPostgres(
	host, port, user, password, dbname string,
) (*sql.DB, error) {

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslorprmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	logger.Log.Infow("PostgreSQL connected", "host", host, "db", dbname)

	return db, nil
}
