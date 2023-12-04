package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp"
	"code.cestc.cn/ccos/cnm/ops-base/trace"
	baseutils "code.cestc.cn/ccos/cnm/ops-base/utils"
	"code.cestc.cn/ccos/cnm/ops-base/utils/idutils"
	"code.cestc.cn/ccos/cnm/ops-base/utils/userutils"
)

var (
	timeFormatLayout = "2006-01-02T15:04:05.000Z"
)

const (
	ResourceId               = "resourceId"
	ResourceDisplayName      = "resourceDisplayName"
	ResourceType             = "resourceType"
	ResourceTypeId           = "resourceTypeId"
	ResourceGroupId          = "resourceGroupId"
	ResourceGroupDisplayName = "resourceGroupDisplayName"
	DepartmentIdDisplayName  = "departmentIdDisplayName"
	TenantDisplayName        = "tenantDisplayName"
)

const (
	// DEBUG 调试级别消息
	// INFO 信息性事件
	// NOTICE 正常但重要的事件
	// WARN 警告情况
	// ERROR 非常严重错误状况
	// CRIT 临界情况
	// ALERT 必须立即采取措施
	// EMERG 系统不可用，程序将不可用
	DEBUG  EventLevel = "DEBUG"
	INFO   EventLevel = "INFO"
	NOTICE EventLevel = "NOTICE"
	WARN   EventLevel = "WARN"
	ERROR  EventLevel = "ERROR"
	CRIT   EventLevel = "CRIT"
	ALERT  EventLevel = "ALERT"
	EMERG  EventLevel = "EMERG"

	CREATE  ActionType = "CREATE"
	UPDATE  ActionType = "UPDATE"
	DELETE  ActionType = "DELETE"
	GET     ActionType = "GET"
	LIST    ActionType = "LIST"
	IMPORT  ActionType = "IMPORT"
	EXPORT  ActionType = "EXPORT"
	OPERATE ActionType = "OPERATE"
)

const (
	OpLogUserKey     = "op-log-user"
	OpLogReplyKey    = "op-log-reply"
	OpLogSwitchKey   = "op-log-switch"
	OpLogSwitchOpen  = 1 // 开
	OpLogSwitchClose = 2 // 关
)

const (
	requestFrom = "x-request-from"

	Openapi = "openapi"
	Webapi  = "webapi"

	RequestFromOm     = "om"     // 运维侧
	RequestFromTenant = "tenant" // 租户运营侧
)

type EventLevel string
type ActionType string

type LogConf struct {
	ServiceCode        string     // 由产品统一定义的云服务名称，需与云资源模型中定义的产品编码一致
	ServiceDisplayName string     // 由产品统一定义的云服务名称，需与云资源模型中定义的产品编码一致
	UserSecretFile     string     // 用户信息解密秘钥
	ActionCode         string     // 自行或由产品统一定义的云服务事件名称，需与云资源模型中定义的action一致
	ActionDisplayName  string     // 事件名称
	ActionType         ActionType // 根据当前接口的操作类型，定义当前接口是读接口还是写接口（ Write,Read）
	EventLevel         EventLevel // 等级
	RequestRegion      string     // 当前服务发起的地域，如服务部署在北京则 cn-beijing
	Extended           string     // 扩展
	RequestDescription string
}

// OperatorLog 操作日志
func OperatorLog(cfg LogConf) gin.HandlerFunc {
	return func(c *gin.Context) {

		if GetOpLogSwitch(c) == OpLogSwitchClose || !isOutApi(c) {
			c.Next()
		} else {
			start := time.Now()

			// response write
			writer := &responseWriter{
				c.Writer,
				bytes.NewBuffer([]byte{}),
			}
			c.Writer = writer

			// request
			var reqBody string
			if baseutils.InStringArray(c.Request.Method, []string{http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodPut}) {
				requestBody, _ := c.GetRawData()
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
				reqBody = string(requestBody)
			}

			c.Next()

			// 返回数据
			var replyBody = writer.b.String()
			var reply opshttp.WrapResp
			if len(replyBody) > 0 {
				_ = jsoniter.Unmarshal([]byte(replyBody), &reply)
			}

			if len(reqBody) >= 512 {
				reqBody = reqBody[0:512]
			}

			if len(replyBody) >= 512 {
				replyBody = replyBody[0:512]
			}

			// 请求id
			traceId := trace.ExtraTraceID(c)

			// 兜底，trace获取不到，生成
			if len(traceId) <= 0 {
				traceId = idutils.GetUUID()
			}

			eventSource := c.Request.Header.Get("Origin")

			if len(eventSource) <= 0 {
				eventSource = c.GetHeader("Host")
			}
			if len(eventSource) <= 0 {
				eventSource = c.GetHeader("Referer")
			}

			if len(eventSource) > 0 {
				split := strings.Split(eventSource, "//")
				if len(split) == 2 {
					eventSource = strings.Split(split[1], "/")[0]
				}
			}

			// 结束时间
			end := time.Now()

			// 获取用户信息
			// 先从ctx中获取，如果没有，再解密获取
			user := userutils.GetUserByContext(c)
			if len(user.GetSystem()) <= 0 {
				user, _ = userutils.GetUser(c.Request, cfg.UserSecretFile)
			}

			// 用户信息（兼容sso，未登录没有用户信息）
			userId, userCode, system := user.GetUserId(), user.GetUserCode(), user.GetSystem()
			if len(user.GetSystem()) <= 0 {
				opUser := GetOpLogUser(c)
				userId = opUser.UserId
				userCode = opUser.UserCode
				system = opUser.System
			}

			// 获取response
			newReply := GetOpLogReply(c)
			if len(newReply.Code) > 0 && len(newReply.Msg) > 0 {
				reply = newReply
			}

			// 状态
			status := getStatus(c.Writer.Status(), reply.Code)

			log := logging.WithOpFields(
				zap.String("eventId", idutils.GetUUID()),
				zap.String("eventVersion", "1"),
				zap.String("eventSource", eventSource),
				zap.String("sourceIpAddress", getRealIp(c)),
				zap.String("userAgent", c.Request.UserAgent()),
				zap.String("serviceDisplayName", cfg.ServiceDisplayName),
				zap.String("actionCode", cfg.ActionCode),
				zap.String("actionDisplayName", cfg.ActionDisplayName),
				zap.String("actionType", string(cfg.ActionType)),
				zap.String("level", string(cfg.EventLevel)),
				zap.String("requestTime", start.UTC().Format(timeFormatLayout)),
				zap.String("requestStartTime", start.UTC().Format(timeFormatLayout)),
				zap.String("requestEndTime", end.UTC().Format(timeFormatLayout)),
				zap.String("requestRegion", cfg.RequestRegion),
				zap.String("resourceDisplayName", c.GetString(ResourceDisplayName)),
				zap.String("resourceId", c.GetString(ResourceId)),
				zap.String("resourceType", c.GetString(ResourceType)),
				zap.String("resourceTypeId", c.GetString(ResourceTypeId)),
				zap.String("requestParameters", reqBody),
				zap.String("result", status),
				zap.String("responseElements", replyBody),
				zap.String("traceInfo", traceId),
				zap.String("errorCode", reply.Code),
				zap.String("errorMessage", reply.Msg),
				zap.String("error_message_cn", ""),
				zap.String("error_message_en", ""),
				zap.String("userCode", userCode),
				zap.String("userId", userId),
				zap.String("userDisplayName", ""),
			)
			key := "[OP_ACTION_TRAIL_LOG]"

			if system == userutils.SystemOps {
				log = log.With(
					zap.String("serviceCode", cfg.ServiceCode),
					zap.String("resourceGroupId", c.GetString(ResourceGroupId)),
					zap.String("resourceGroupDisplayName", c.GetString(ResourceGroupDisplayName)),
					zap.String("departmentId", user.GetDepartmentId()),
					zap.String("departmentIdDisplayName", c.GetString(DepartmentIdDisplayName)),
					zap.String("tenantId", user.GetTenantId()),
					zap.String("tenantDisplayName", c.GetString(TenantDisplayName)),
					zap.String("requestUrl", c.Request.URL.Path),
					zap.String("requestDescription", cfg.RequestDescription),
					zap.String("extended", cfg.Extended),
				)
				key = "[TENANT_OP_ACTION_TRAIL_LOG]"
			}

			log.OpLogPrint(key)
		}
	}
}

func getStatus(status int, code string) string {
	b, _ := regexp.MatchString("^2[0-9]{2}", strconv.Itoa(status))
	if b || strings.ToLower(code) == "success" {
		return "Success"
	}
	return "Fail"
}

func getRealIp(c *gin.Context) string {
	ip := c.GetHeader("x-forwarded-for")
	if ip == "" || ip == "unknown" {
		ip = c.GetHeader("Proxy-Client-IP")
		if ip == "" || ip == "unknown" {
			ip = c.GetHeader("WL-Proxy-Client-IP")
			if ip == "" || ip == "unknown" {
				ip = c.ClientIP()
			}
		}
	}
	return strings.Split(ip, ",")[0]
}

type OpLogUser struct {
	UserId   string
	UserCode string
	System   string
}

func SetOpLogUser(c *gin.Context, user OpLogUser) {
	c.Set(OpLogUserKey, user)
}

func GetOpLogUser(c *gin.Context) OpLogUser {
	value, _ := c.Get(OpLogUserKey)
	user, _ := value.(OpLogUser)
	return user
}

func SetOpLogReply(c *gin.Context, reply opshttp.WrapResp) {
	c.Set(OpLogReplyKey, reply)
}

func GetOpLogReply(c *gin.Context) opshttp.WrapResp {
	value, _ := c.Get(OpLogReplyKey)
	reply, _ := value.(opshttp.WrapResp)
	return reply
}

func SetOpLogSwitch(c *gin.Context, logSwitch int64) {
	c.Set(OpLogSwitchKey, logSwitch)
}

func GetOpLogSwitch(c *gin.Context) int64 {
	value, _ := c.Get(OpLogSwitchKey)
	reply, _ := value.(int64)
	return reply
}

func isOutApi(c *gin.Context) bool {
	reqFrom := c.GetHeader(requestFrom)
	if len(reqFrom) <= 0 {
		return false
	}

	requestFromSplit := strings.Split(reqFrom, "/")
	if len(requestFromSplit) < 2 {
		return false
	}

	if !baseutils.InStringArray(requestFromSplit[0], []string{Openapi, Webapi}) {
		return false
	}

	if !baseutils.InStringArray(requestFromSplit[1], []string{RequestFromOm, RequestFromTenant}) {
		return false
	}

	return true
}
