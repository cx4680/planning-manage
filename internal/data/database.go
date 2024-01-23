package data

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/app/settings"
)

const (
	maxIdleConn = 3
	maxOpenConn = 5
	maxLifeTime = 600
	maxIdleTime = 60
)

var (
	// DB 数据库链接单例。
	DB *gorm.DB
)

func InitDatabase(setting *settings.Setting) {
	// 初始化GORM日志配置
	var connString = setting.MySQLDSN
	if !setting.MySQLInsecure {
		connString = mySQLConnString(setting)
	}
	log.Infof("database config: %s", connString)
	db, err := gorm.Open(mysql.Open(connString), &gorm.Config{
		QueryFields: true,
		Logger:      logger.Default.LogMode(logger.Info),
	})

	if connString == "" || err != nil {
		log.Error(err, "Open database error")
		panic(err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		log.Error(err, "MySQL lost")
		panic(err)
	}

	// 设置连接池
	// 空闲
	sqlDB.SetMaxIdleConns(maxIdleConn)
	// 打开
	sqlDB.SetMaxOpenConns(maxOpenConn)
	// 设置连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifeTime) * time.Second)
	// 设置空闲连接最大保持时间
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)
	DB = db

	// 数据库迁移
	// migrateTable()
	// 数据初始化
	migrateData(connString)
}

func migrateData(dsn string) error {
	_ = &file.File{}
	db, err := sql.Open("mysql", dsn+"&multiStatements=true")
	if err != nil {
		log.Error(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			log.Error(err)
		}
	}(db)
	if err != nil {
		return err
	}
	instance, err := migratemysql.WithInstance(db, &migratemysql.Config{
		MigrationsTable: "planning_manage_migrations",
	})
	if err != nil {
		log.Errorf("MigrationsTable err, schema migration message: %s", err.Error())
		return err
	}
	defer func(instance database.Driver) {
		err = instance.Close()
		if err != nil {
			log.Errorf("instance close %v", err)
		}
	}(instance)
	m, err := migrate.NewWithDatabaseInstance("file://./migrations",
		"planning_manage", instance)
	if err != nil {
		log.Errorf("NewWithDatabaseInstance message: %s", err.Error())
		return err
	}
	m.Log = new(migrateLog)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Errorf("schema migration message: %s", err.Error())
		return err
	}
	return nil
}

type migrateLog struct {
}

func (l *migrateLog) Printf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

func (l *migrateLog) Verbose() bool {
	return true
}

func PaginateDB(context *gin.Context) *gorm.DB {
	size := context.GetInt(constant.Size)
	current := context.GetInt(constant.Current)
	offset := (current - 1) * size
	if offset < 0 {
		offset = 0
	}
	return DB.Offset(offset).Limit(size)
}

func Paginate(current int, size int) *gorm.DB {
	offset := (current - 1) * size
	if offset < 0 {
		offset = 0
	}
	return DB.Offset(offset).Limit(size)
}

func DBColumnEqualFunc(db *gorm.DB, column string, value interface{}) {
	DBColumnConditionEqualFunc(column != "", db, column, value)
}

func DBColumnNotEqualFunc(db *gorm.DB, column string, value string) {
	DBColumnConditionNotEqualFunc(column != "", db, column, value)
}

func DBColumnConditionNotEqualFunc(condition bool, db *gorm.DB, column string, value string) {
	stringDecode := StringDecode(value)
	if condition && column != "" && stringDecode != "" {
		db.Where(fmt.Sprintf("%s != ? ", column), stringDecode)
	}
}

func DBColumnConditionEqualFunc(condition bool, db *gorm.DB, column string, value interface{}) {
	switch value.(type) {
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		if condition && column != "" {
			db.Where(fmt.Sprintf("%s = ? ", column), value)
		}
		break
	default:
		newValue, _ := json.Marshal(value)
		if condition && column != "" && string(newValue) != "" {
			db.Where(fmt.Sprintf("%s = ? ", column), strings.Trim(string(newValue), `"`))
		}
		break
	}
}

func DBColumnLikeFunc(db *gorm.DB, column string, value string) {
	DBColumnConditionLikeFunc(column != "", db, column, value)
}

func DBColumnConditionLikeFunc(condition bool, db *gorm.DB, column string, value string) {
	stringDecode := StringDecode(value)
	if condition && column != "" && stringDecode != "" {
		db.Where(fmt.Sprintf("%s like %s", column, "CONCAT('%',?,'%') "), stringDecode)
	}
}

func DBColumnInFunc(db *gorm.DB, column string, values []string) {
	DBColumnConditionInFunc(column != "", db, column, values)
}

func DBColumnConditionInFunc(condition bool, db *gorm.DB, column string, values []string) {
	stringArrayDecode := StringArrayDecode(values)
	if condition && column != "" && len(stringArrayDecode) != 0 {
		db.Where(fmt.Sprintf("%s in ? ", column), stringArrayDecode)
	}
}

func DBColumnOrderFunc(db *gorm.DB, column string, value string) {
	DBColumnConditionOrderFunc(column != "", db, column, value)
}

func DBColumnConditionOrderFunc(condition bool, db *gorm.DB, column string, value string) {
	if condition && column != "" && value != "" {
		db.Order(fmt.Sprintf("%s %s", column, value))
	}
}

func ErrRecordNotFound(db *gorm.DB) bool {
	return errors.Is(db.Error, gorm.ErrRecordNotFound) || db.RowsAffected == 0
}

// mySQLConnString mySQL secret encryption
func mySQLConnString(setting *settings.Setting) string {
	mySQLPwd := setting.MySQLDBPassword
	if mySQLPwd == "" {
		log.Info("Mysql password is empty")
		return ""
	}
	return fmt.Sprintf("%s:%s%s", setting.MySQLUser, mySQLPwd, os.Getenv(settings.EnvMySQLDsnOptions))
}
