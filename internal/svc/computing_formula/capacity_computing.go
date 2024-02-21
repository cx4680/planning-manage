package computing_formula

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"math"
	"strconv"
	"strings"
)

func SpecialCapacityComputing(serverCapacityMap map[int64]float64, capConvertBaselineMap map[int64]*entity.CapConvertBaseline) map[string]float64 {
	//按产品将容量输入参数分类
	var productCapMap = make(map[string][]*entity.CapConvertBaseline)
	for _, v := range capConvertBaselineMap {
		productCapMap[v.ProductCode] = append(productCapMap[v.ProductCode], v)
	}
	var capActualResMap = make(map[string]float64)
	for k, v := range productCapMap {
		switch k {
		case "CKE":
			var vCpu, memory, cluster float64
			for _, capConvertBaseline := range v {
				switch capConvertBaseline.CapPlanningInput {
				case "vCPU":
					vCpu = serverCapacityMap[capConvertBaseline.Id]
				case "内存":
					memory = serverCapacityMap[capConvertBaseline.Id]
				case "容器集群数":
					cluster = serverCapacityMap[capConvertBaseline.Id]
				}
			}
			cpuCapActualRes := 48*cluster + 16*vCpu/0.7/14.6
			memoryCapActualRes := 96*cluster + 32*memory/0.7/29.4
			capActualResMap["ECS_VCPU"] = cpuCapActualRes
			capActualResMap["ECS_MEM"] = memoryCapActualRes
		default:
			//serverNumber = General(number, featureNumber, capActualResBaseline, capServerCalcBaseline, serverBaseline)
		}
	}
	return capActualResMap
}

func CapacityComputing(number, featureNumber int, capActualResBaseline *entity.CapActualResBaseline, capServerCalcBaseline *entity.CapServerCalcBaseline, serverBaseline *entity.ServerBaseline, specialCapActualResMap map[string]float64) int {
	//总消耗
	capActualResNumber := CapActualRes(number, featureNumber, capActualResBaseline)
	switch capServerCalcBaseline.ExpendNodeRoleCode {
	case "COMPUTE":
		capActualResNumber += specialCapActualResMap[capActualResBaseline.ExpendResCode]
	}
	//单个服务器消耗
	capServerCalcNumber := CapServerCalc(capActualResBaseline.ExpendResCode, capServerCalcBaseline, serverBaseline)
	//总消耗除以单个服务器消耗，等于服务器数量
	serverNumber := math.Ceil(capActualResNumber / capServerCalcNumber)
	return int(serverNumber)
}

func CapActualRes(number, featureNumber int, capActualResBaseline *entity.CapActualResBaseline) float64 {
	if featureNumber <= 0 {
		featureNumber = 1
	}
	numerator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioNumerator, 64)
	if numerator == 0 {
		numerator = float64(featureNumber)
	}
	denominator, _ := strconv.ParseFloat(capActualResBaseline.OccRatioDenominator, 64)
	if denominator == 0 {
		denominator = float64(featureNumber)
	}
	//总消耗
	return float64(number) / numerator * denominator
}

func CapServerCalc(expendResCode string, capServerCalcBaseline *entity.CapServerCalcBaseline, serverBaseline *entity.ServerBaseline) float64 {
	//判断用哪个容量参数
	var singleCapacity int
	if strings.Contains(expendResCode, "_VCPU") {
		singleCapacity = serverBaseline.Cpu
	}
	if strings.Contains(expendResCode, "_MEM") {
		singleCapacity = serverBaseline.Memory
	}
	if strings.Contains(expendResCode, "_DISK") {
		singleCapacity = serverBaseline.StorageDiskNum * serverBaseline.StorageDiskCapacity
	}

	nodeWastage, _ := strconv.ParseFloat(capServerCalcBaseline.NodeWastage, 64)
	waterLevel, _ := strconv.ParseFloat(capServerCalcBaseline.WaterLevel, 64)
	//单个服务器消耗
	if capServerCalcBaseline.NodeWastageCalcType == 1 {
		return (float64(singleCapacity) - nodeWastage) * waterLevel
	} else {
		return (float64(singleCapacity) * (1 - nodeWastage)) * waterLevel
	}
}

func EcsCapacityComputing(cpu, memory, count, featureNumber int, capPlanningInput, arch string, capActualResBaseline *entity.CapActualResBaseline, capServerCalcBaseline *entity.CapServerCalcBaseline, serverBaseline *entity.ServerBaseline, specialCapActualResMap map[string]float64) int {
	var number int
	if capPlanningInput == "vCPU" {
		number = cpu * count
	}
	if capPlanningInput == "内存" {
		number = 138 + 8 + 16 + 8*cpu*count + memory*count/512
		if arch == "ARM" {
			number += 128
		}
		number = number / 1024
		featureNumber = 1
	}
	return CapacityComputing(number, featureNumber, capActualResBaseline, capServerCalcBaseline, serverBaseline, specialCapActualResMap)
}
