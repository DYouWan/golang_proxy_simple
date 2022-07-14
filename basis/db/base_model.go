package db

import (
	"database/sql"
	"time"
)

type BaseModel struct {
	ID          string `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	InvalidSign int            `sql:"type:int(1);not null;default:0"`
	OperationID sql.NullString `sql:"type:varchar(200)"`
}
