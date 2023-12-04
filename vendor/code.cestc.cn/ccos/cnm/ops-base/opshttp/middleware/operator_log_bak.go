package middleware

//
//var (
//	timeFormatLayout = "2006-01-02T15:04:05.000Z"
//)
//
//const (
//	ResourceId               = "resourceId"
//	ResourceDisplayName      = "resourceDisplayName"
//	ResourceType             = "resourceType"
//	ResourceTypeId           = "resourceTypeId"
//	ResourceGroupId          = "resourceGroupId"
//	ResourceGroupDisplayName = "resourceGroupDisplayName"
//)
//
//type OperatorLogConfig struct {
//	ServiceCode        string              `yaml:"serviceCode"`
//	ServiceDisplayName string              `yaml:"serviceDisplayName"`
//	UserSecretFile     string              `yaml:"userSecretFile"`
//	ActionList         []OperatorLogAction `yaml:"actionList"`
//}
//
//type OperatorLogAction struct {
//	Url               string `yaml:"url"`
//	ActionCode        string `yaml:"actionCode"`
//	ActionDisplayName string `yaml:"actionDisplayName"`
//	ActionType        string `yaml:"actionType"`
//	EventLevel        string `yaml:"eventLevel"`
//}
//
//// 操作日志
//func loggingOperator(cfg OperatorLogConfig) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		start := time.Now()
//		c.Next()
//		if len(cfg.ActionList) > 0 {
//			action := OperatorLogAction{}
//			for _, v := range cfg.ActionList {
//				if v.Url == c.Request.URL.Path {
//					action = v
//				}
//			}
//			if len(action.ActionCode) > 0 {
//
//				var reply opshttp.WrapResp
//				replyStr := c.GetString(respKey)
//				if len(replyStr) > 0 {
//					_ = jsoniter.Unmarshal([]byte(replyStr), &reply)
//				}
//
//				// 请求id
//				traceId := trace.ExtraTraceID(c)
//				eventSource := c.Request.Header.Get("Origin")
//
//				if len(eventSource) <= 0 {
//					eventSource = c.GetHeader("Host")
//				}
//				if len(eventSource) <= 0 {
//					eventSource = c.GetHeader("Referer")
//				}
//
//				if len(eventSource) > 0 {
//					split := strings.Split(eventSource, "//")
//					if len(split) == 2 {
//						eventSource = strings.Split(split[1], "/")[0]
//					}
//				}
//
//				// 状态
//				status := getStatus(c.Writer.Status())
//
//				// 结束时间
//				end := time.Now()
//
//				// 请求来源
//				reqFrom := c.GetHeader("x-request-from")
//
//				// 获取用户信息
//				user, _ := userutils.GetUser(c.Request, cfg.UserSecretFile)
//
//				var key string
//				if len(reqFrom) > 0 {
//					requestFromSplitOm := strings.Split(reqFrom, "/")
//					if len(requestFromSplitOm) >= 2 {
//						if requestFromSplitOm[1] == "om" {
//							key = "[OP_ACTION_TRAIL_LOG]"
//						} else {
//							key = "[TENANT_OP_ACTION_TRAIL_LOG]"
//						}
//					}
//				}
//
//				logging.OpLogPrint(key,
//					zap.String("eventId", traceId),
//					zap.String("eventVersion", "1"),
//					zap.String("eventSource", eventSource),
//					zap.String("sourceIpAddress", utils.GetHost()),
//					zap.String("userAgent", c.Request.UserAgent()),
//					zap.String("serviceDisplayName", cfg.ServiceDisplayName),
//					zap.String("actionCode", action.ActionCode),
//					zap.String("actionDisplayName", action.ActionDisplayName),
//					zap.String("actionType", action.ActionType),
//					zap.String("level", action.EventLevel),
//					zap.String("requestTime", start.UTC().Format(timeFormatLayout)),
//					zap.String("requestStartTime", start.UTC().Format(timeFormatLayout)),
//					zap.String("requestEndTime", end.UTC().Format(timeFormatLayout)),
//					zap.String("requestRegion", os.Getenv("REGION")),
//					zap.String("resourceDisplayName", c.GetString(ResourceDisplayName)),
//					zap.String("resourceId", c.GetString(ResourceId)),
//					zap.String("resourceType", c.GetString(ResourceType)),
//					zap.String("resourceTypeId", c.GetString(ResourceTypeId)),
//					zap.String("resourceGroupId", c.GetString(ResourceGroupId)),
//					zap.String("resourceGroupDisplayName", c.GetString(ResourceGroupDisplayName)),
//					zap.String("requestParameters", c.GetString(reqKey)),
//					zap.String("result", status),
//					zap.String("responseElements", replyStr),
//					zap.String("traceInfo", traceId),
//					zap.String("errorCode", reply.Code),
//					zap.String("errorMessage", reply.Msg),
//					zap.String("error_message_cn", ""),
//					zap.String("error_message_en", ""),
//					zap.String("userCode", user.GetUserCode()),
//					zap.String("userId", user.GetUserId()),
//					zap.String("userDisplayName", ""),
//				)
//			}
//		}
//	}
//}
//
//func getStatus(status int) string {
//	b, _ := regexp.MatchString("^2[0-9]{2}", strconv.Itoa(status))
//	if b {
//		return "Success"
//	}
//	return "Fail"
//}
