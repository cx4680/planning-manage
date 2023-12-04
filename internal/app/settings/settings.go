package settings

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/acmestack/godkits/gox/stringsx"
	"github.com/gin-gonic/gin"

	_ "github.com/joho/godotenv/autoload"

	"code.cestc.cn/zhangzhi/planning-manage/internal/api/constant"
	"code.cestc.cn/zhangzhi/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/result"
)

const (
	EnvMySQLDsnOptions = "MYSQL_DSN_OPTIONS"
	EnvMySQLUser       = "MYSQL_USER"
	EnvMySQLDBPassword = "MYSQL_DB_PASSWORD"
	EnvMySQLInsecure   = "MYSQL_INSECURE"
	EnvMySQLDSN        = "MYSQL_DSN"
	EnvNamespace       = "NAMESPACE"
)

type Setting struct {
	LogLevel               string
	LogPath                string
	Https                  bool
	Port                   string
	Writer                 io.Writer
	EnvGinMode             string
	MySQLDSN               string
	MySQLUser              string
	MySQLDBPassword        string
	CustomRecoveryGinError func(context *gin.Context)
	MySQLInsecure          bool
	Namespace              string
	HttpCallTimeoutMinute  time.Duration
}

func NewSetting() *Setting {
	s := &Setting{
		LogLevel:        stringsx.DefaultIfEmpty(os.Getenv("LOG_LEVEL"), "info"),
		LogPath:         stringsx.DefaultIfEmpty(os.Getenv("LOG_PATH"), "/var/log/planning-manage.log"),
		Https:           stringsx.DefaultIfEmpty(os.Getenv("HTTPS"), "true") == "true",
		Port:            stringsx.DefaultIfEmpty(os.Getenv("PORT"), "8443"),
		Writer:          io.MultiWriter(os.Stdout),
		EnvGinMode:      stringsx.DefaultIfEmpty(os.Getenv(gin.EnvGinMode), "debug"),
		MySQLUser:       stringsx.DefaultIfEmpty(os.Getenv(EnvMySQLUser), "root"),
		MySQLDBPassword: stringsx.DefaultIfEmpty(os.Getenv(EnvMySQLDBPassword), "123456"),
		MySQLInsecure:   stringsx.DefaultIfEmpty(os.Getenv(EnvMySQLInsecure), "false") == "true",
		MySQLDSN:        stringsx.DefaultIfEmpty(os.Getenv(EnvMySQLDSN), "root:123456@tcp(mysql-planning-manage-svc.planning-manage:3306)/planning_manage?timeout=10s&readTimeout=10s&writeTimeout=10s&parseTime=true&loc=Local&charset=utf8mb4,utf8"),
		CustomRecoveryGinError: func(context *gin.Context) {
			result.Failure(context, errorcodes.UnknownError, http.StatusInternalServerError)
		},
		Namespace:             stringsx.DefaultIfEmpty(os.Getenv(EnvNamespace), constant.NameSpace),
		HttpCallTimeoutMinute: 3 * time.Minute,
	}

	// MySQLDSN := fmt.Sprintf("%s:%s%s", os.Getenv(EnvMySQLUser), os.Getenv(EnvMySQLDBPassword), os.Getenv(EnvMySQLDsnOptions))
	// s.MySQLDSN = MySQLDSN

	return s
}
