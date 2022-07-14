package migrations

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

//Bootstrap 创建迁移表,保留已经运行的数据库迁移
func Bootstrap(db *gorm.DB) error {
	migrationName := "bootstrap_migrations"
	migration := &Migration{Name: migrationName}
	// 创建migrations表
	if err := db.CreateTable(new(Migration)).Error; err != nil {
		return fmt.Errorf("创建migrations表失败: %s", err)
	}
	// 在migrations表中添加记录
	if err := db.Create(migration).Error; err != nil {
		return fmt.Errorf("将记录保存到迁移表中出错: %s", err)
	}
	return nil
}
