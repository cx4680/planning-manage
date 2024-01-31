package software_count

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"fmt"
)

func SoftwareCount(softwareBomLicenseBaseline *entity.SoftwareBomLicenseBaseline, cloudProductPlanning *entity.CloudProductPlanning, serverPlanning *entity.ServerPlanning, serverBaseline *entity.ServerBaseline, serverCapPlanningMap map[string]*entity.ServerCapPlanning) int {
	switch softwareBomLicenseBaseline.ServiceCode {
	case "ECS", "BMS", "VPC":
		return serverPlanning.Number * serverBaseline.CpuNum
	case "CKE":
		serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", "CKE", "vCPU")]
		return serverCapPlanning.Number / 100
	case "EBS", "EFS", "OSS", "CBR":
		return 1
	case "CNBH":
		return 1
	case "CWP":
		return 1
	case "DES":
		if softwareBomLicenseBaseline.SellType == "升级维保" {
			return cloudProductPlanning.ServiceYear
		}
		return 1
	case "KAFKA", "ROCKETMQ", "APIM":
		return 1
	case "CSP":
		return 1
	case "CONNECT":
		return 1
	default:
		return 1
	}
}

const (
	PlatformBom     = "0100115148387809"
	SoftwareBaseBom = "0100115150861886"
)

var ServiceYearBom = map[int]string{1: "0100115152958526", 2: "0100115153975617", 3: "0100115154780568", 4: "0100115155303482", 5: "0100115156784743"}
