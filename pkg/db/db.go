package db

import (
	"database/sql"

	"github.com/ElfAstAhe/go-service-template/pkg/helper"
)

type DB interface {
	GetDriver() string
	GetDB() *sql.DB
	GetDSN() string
	GetHelper() helper.DBHelper
	Close() error
}
