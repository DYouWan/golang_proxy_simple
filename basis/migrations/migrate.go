package migrations

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"iotWeb/basis/logging"
)

// Migration represents a single database migration
type Migration struct {
	gorm.Model
	Name string `sql:"size:255"`
}

// MigrationStage ...
type MigrationStage struct {
	Name     string
	Function func(db *gorm.DB, name string) error
}

// Migrate ...
func Migrate(db *gorm.DB, migrations []MigrationStage) error {
	for _, m := range migrations {
		if MigrationExists(db, m.Name) {
			logging.INFO.Printf("跳过 %s 迁移",  m.Name)
			continue
		}
		logging.INFO.Printf("开始 %s 迁移", m.Name)

		if err := m.Function(db, m.Name); err != nil {
			return err
		}
		if err := SaveMigration(db, m.Name); err != nil {
			return err
		}
		logging.INFO.Printf("%s 迁移成功", m.Name)
	}
	return nil
}

// MigrateAll runs bootstrap, then all migration functions listed against
// the specified database and logs any errors
func MigrateAll(db *gorm.DB, migrationFunctions []func(*gorm.DB) error) {
	if err := Bootstrap(db); err != nil {
		logging.ERROR.Print(err)
	}

	for _, m := range migrationFunctions {
		if err := m(db); err != nil {
			logging.ERROR.Print(err)
		}
	}
}

// MigrationExists 判断表中是否已经有名称为migrationName的迁移记录，如果有则跳过迁移，否则执行迁移
func MigrationExists(db *gorm.DB, migrationName string) bool {
	migration := new(Migration)
	return !db.Where("name = ?", migrationName).First(migration).RecordNotFound()
}

// SaveMigration 将迁移记录保存到迁移表
func SaveMigration(db *gorm.DB, migrationName string) error {
	migration := &Migration{Name: migrationName}

	if err := db.Create(migration).Error; err != nil {
		return fmt.Errorf("保存迁移记录到迁移表时报错 %s", err)
	}
	return nil
}
