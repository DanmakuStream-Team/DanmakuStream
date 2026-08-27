//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// OpenTemporaryMySQL creates an isolated database and removes it after the test.
func OpenTemporaryMySQL(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	adminDSN := os.Getenv("DANMAKU_TEST_ADMIN_DSN")
	if adminDSN == "" {
		t.Skip("set DANMAKU_TEST_ADMIN_DSN to run MySQL integration tests")
	}
	adminConfig, err := mysqldriver.ParseDSN(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig.DBName = ""
	adminDB, err := sql.Open("mysql", adminConfig.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	databaseName := fmt.Sprintf("danmakustream_integration_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(context.Background(), "CREATE DATABASE `"+databaseName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		_ = adminDB.Close()
		t.Fatal(err)
	}

	testConfig := *adminConfig
	testConfig.DBName = databaseName
	db, err := gorm.Open(gormmysql.Open(testConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		_, _ = adminDB.ExecContext(context.Background(), "DROP DATABASE `"+databaseName+"`")
		_ = adminDB.Close()
		t.Fatal(err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_, _ = adminDB.ExecContext(context.Background(), "DROP DATABASE `"+databaseName+"`")
		_ = adminDB.Close()
	})
	return db
}
