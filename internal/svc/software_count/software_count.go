package software_count

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"fmt"
	"math"
)

func SoftwareCount(i int, softwareBomLicenseBaseline *entity.SoftwareBomLicenseBaseline, cloudProductPlanning *entity.CloudProductPlanning, serverPlanning *entity.ServerPlanning, serverBaseline *entity.ServerBaseline, serverCapPlanningMap map[string]*entity.ServerCapPlanning) int {
	switch softwareBomLicenseBaseline.ServiceCode {
	case "ECS", "BMS", "VPC":
		return serverPlanning.Number * serverBaseline.CpuNum
	case "CKE":
		serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "vCPU")]
		if serverCapPlanning != nil {
			return serverCapPlanning.Number / 100
		}
		return 1
	case "EBS", "EFS", "OSS", "CBR":
		serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "容量")]
		if serverCapPlanning != nil && serverCapPlanning.Features == "三副本" {
			return int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 1 / 3))
		}
		return int(math.Ceil(float64(serverPlanning.Number*serverBaseline.StorageDiskNum*serverBaseline.StorageDiskCapacity) / 1024 * 0.9 * 0.91 * 2 / 3))
	case "CNFW", "CWAF", "CNBH", "CWP", "DES":
		if softwareBomLicenseBaseline.SellType == "升级维保" {
			return cloudProductPlanning.ServiceYear
		}
		return 1
	case "KAFKA", "ROCKETMQ", "APIM", "CONNECT":
		if i > 1 {
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "vCPU")]
			if serverCapPlanning != nil && serverCapPlanning.Number > 200 {
				return int(math.Ceil(float64(serverCapPlanning.Number-200) / 100))
			}
			return 0
		}
		return 1
	case "CSP":
		if i > 1 {
			serverCapPlanning := serverCapPlanningMap[fmt.Sprintf("%v-%v", softwareBomLicenseBaseline.ServiceCode, "vCPU")]
			if serverCapPlanning != nil && serverCapPlanning.Number > 500 {
				return int(math.Ceil(float64(serverCapPlanning.Number-500) / 100))
			}
			return 0
		}
		return 1
	default:
		return 1
	}
}

const (
	PlatformName    = "平台规模授权"
	PlatformCode    = "Platform"
	PlatformBom     = "0100115148387809"
	SoftwareName    = "软件base"
	SoftwareCode    = "SoftwareBase"
	SoftwareBaseBom = "0100115150861886"
	ServiceYearName = "平台升级维保"
	ServiceYearCode = "ServiceYear"
)

var ServiceYearBom = map[int]string{1: "0100115152958526", 2: "0100115153975617", 3: "0100115154780568", 4: "0100115155303482", 5: "0100115156784743"}
