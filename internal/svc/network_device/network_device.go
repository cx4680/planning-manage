package network_device

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"code.cestc.cn/ccos/common/planning-manage/internal/svc/software_bom"

	"github.com/xuri/excelize/v2"

	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/user"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/datetime"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/ip_demand"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/plan"
)

func GetDevicePlanByPlanId(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	// 根据方案ID查询网络设备规划
	devicePlan, err := SearchDevicePlanByPlanId(planId)
	if err != nil {
		log.Errorf("[searchDevicePlanByPlanId] search device plan by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if devicePlan.PlanId == 0 {
		// 说明第一次进入，赋默认值方便前端展示
		devicePlan.DeviceType = 0
		devicePlan.ApplicationDispersion = "1"
		devicePlan.AwsServerNum = 44
		devicePlan.AwsBoxNum = 4
		devicePlan.Ipv6 = "0"
		devicePlan.NetworkModel = 1
		devicePlan.TotalBoxNum = 4
		devicePlan.Brand = "华为"
	}
	result.Success(c, devicePlan)
	return
}

func GetBrandsByPlanId(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list network devices bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	planId := request.PlanId
	versionId := request.VersionId
	if versionId == 0 || planId == 0 {
		result.FailureWithMsg(c, errorcodes.InvalidParam, http.StatusBadRequest, errorcodes.ParamError)
		return
	}
	// 根据方案id查询云产品规划信息  取其中一条拿服务器基线表ID
	serverPlanningList, err := queryServerPlanningListByPlanId(planId)
	if err != nil {
		log.Errorf("[QueryServerPlanningListByPlanId] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(serverPlanningList) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.ServerPlanningListEmpty)
		return
	}
	// 根据服务版本id查询网络设备基线表查厂商  去重
	brands, err := GetBrandsByVersionId(versionId)
	if err != nil {
		log.Errorf("[getBrandsByPlanId] search brands by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, brands)
	return
}

func ListNetworkDevices(c *gin.Context) {
	request := &Request{}
	if err := c.BindJSON(&request); err != nil {
		log.Errorf("list network devices bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request); err != nil {
		result.FailureWithMsg(c, errorcodes.InvalidParam, http.StatusBadRequest, err.Error())
		return
	}
	// 入参获取版本ID
	versionId := request.VersionId
	var finalResponse NetworkDevicesResponse
	var response []NetworkDevices
	// 根据方案ID查询网络设备规划表 没有则保存，有则更新
	planId := request.PlanId
	// deviceList, err := SearchDeviceListByPlanId(planId)
	// if err != nil {
	// 	log.Errorf("[searchDeviceListByPlanId] search device list by planId error, %v", err)
	// 	result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
	// 	return
	// }
	// 根据方案id查询服务器规划
	serverPlanningList, err := queryServerPlanningListByPlanId(planId)
	if err != nil {
		log.Errorf("[QueryServerPlanningListByPlanId] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(serverPlanningList) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.ServerPlanningListEmpty)
		return
	}
	// if len(deviceList) > 0 && !request.EditFlag {
	// 	// 不是第一次进入并且也不是编辑网络设备规划 那就不需要重新计算 直接从库里拿
	// 	for _, device := range deviceList {
	// 		networkDevice := NetworkDevices{
	// 			NetworkDeviceRole:     device.NetworkDeviceRole,
	// 			NetworkDeviceRoleName: device.NetworkDeviceRoleName,
	// 			NetworkDeviceRoleId:   device.NetworkDeviceRoleId,
	// 			LogicalGrouping:       device.LogicalGrouping,
	// 			DeviceId:              device.DeviceId,
	// 			Brand:                 device.Brand,
	// 			DeviceModel:           device.DeviceModel,
	// 			ConfOverview:          device.ConfOverview,
	// 			BomId:                 device.BomId,
	// 		}
	// 		// 单独处理下型号列表
	// 		deviceModels, _ := GetModelsByVersionIdAndRoleAndBrand(versionId, device.NetworkDeviceRoleId, request.Brand, request.DeviceType)
	// 		networkDevice.DeviceModels = deviceModels
	// 		response = append(response, networkDevice)
	// 	}
	// 	// 计算网络设备总数
	// 	count := len(response)
	// 	finalResponse.Total = count
	// 	finalResponse.NetworkDeviceList = response
	// 	result.Success(c, finalResponse)
	// 	return
	// }
	// 服务器规划数据转为map
	var nodeRoleServerNumMap = make(map[int64]int)
	for _, value := range serverPlanningList {
		nodeRoleServerNumMap[value.NodeRoleId] = value.Number
	}
	// 根据版本号查询出网络设备角色基线数据
	deviceRoleBaseline, err := SearchDeviceRoleBaselineByVersionId(versionId)
	if err != nil {
		log.Errorf("[searchDeviceRoleBaselineByVersionId] search device role baseline error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(deviceRoleBaseline) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.NetworkDeviceRoleBaselineEmpty)
		return
	}
	response, err = transformNetworkDeviceList(versionId, request, deviceRoleBaseline, nodeRoleServerNumMap)
	if err != nil {
		log.Errorf("[transformNetworkDeviceList] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	// 计算网络设备总数
	total := len(response)
	finalResponse.Total = total
	finalResponse.NetworkDeviceList = response
	devicePlan, err := SearchDevicePlanByPlanId(planId)
	if err != nil {
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if devicePlan.Id == 0 {
		err = CreateDevicePlan(request)
	} else {
		err = UpdateDevicePlan(request, *devicePlan)
	}
	if err != nil {
		log.Errorf("[saveOrUpdateDevicePlan] save or update device plan error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, finalResponse)
	return
}

func SaveDeviceList(c *gin.Context) {
	req := &Request{}
	var networkDeviceList []*entity.NetworkDeviceList
	var ipDemandPlannings []*entity.IpDemandPlanning
	now := datetime.GetNow()
	if err := c.ShouldBind(&req); err != nil {
		log.Errorf("save network devices bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	request := req.Devices
	if len(request) == 0 {
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	planId := req.PlanId
	// 组装网络设备清单数据
	for _, networkDevice := range request {
		networkDeviceList = append(networkDeviceList, &entity.NetworkDeviceList{
			PlanId:                planId,
			NetworkDeviceRole:     networkDevice.NetworkDeviceRole,
			NetworkDeviceRoleId:   networkDevice.NetworkDeviceRoleId,
			NetworkDeviceRoleName: networkDevice.NetworkDeviceRoleName,
			ConfOverview:          networkDevice.ConfOverview,
			LogicalGrouping:       networkDevice.LogicalGrouping,
			DeviceId:              networkDevice.DeviceId,
			Brand:                 networkDevice.Brand,
			DeviceModel:           networkDevice.DeviceModel,
			BomId:                 networkDevice.BomId,
			CreateTime:            now,
			UpdateTime:            now,
			DeleteState:           0,
		})
	}
	deviceList, err := SearchDeviceListByPlanId(planId)
	if err != nil {
		log.Errorf("[searchDeviceListByPlanId] search device list by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	ipDemands, err := ip_demand.SearchIpDemandPlanningByPlanId(planId)
	if err != nil {
		log.Errorf("[SearchIpDemandPlanningByPlanId] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	// 根据云产品版本和云平台类型查询版本ID
	versionId := req.VersionId
	// 根据版本ID查询ip需求基线数据
	ipDemandBaselines, err := ip_demand.GetIpDemandBaselineByVersionId(versionId)
	if err != nil {
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(ipDemandBaselines) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.IpDemandBaselineEmpty)
	}
	userId := user.GetUserId(c)
	if err = data.DB.Transaction(func(tx *gorm.DB) error {
		if len(deviceList) > 0 {
			// 失效库里保存的
			err = ExpireDeviceListByPlanId(tx, planId)
			if err != nil {
				return err
			}
		}
		// 更新方案表的状态
		if err = plan.UpdatePlanStage(tx, planId, constant.PlanStagePlanned, userId, constant.BusinessPlanningEnd, 0); err != nil {
			return err
		}
		// 批量保存网络设备清单
		if err = SaveBatch(tx, networkDeviceList); err != nil {
			return err
		}
		logicGroups, err := GetDeviceRoleLogicGroupByPlanId(tx, planId)
		networkDevicePlan, err := SearchDevicePlanByPlanId(planId)
		if err != nil {
			result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
			return err
		}
		for _, ipDemandBaseline := range ipDemandBaselines {
			for _, logicGroup := range logicGroups {
				if ipDemandBaseline.DeviceRoleId == logicGroup.DeviceRoleId {
					if networkDevicePlan.Ipv6 == constant.Ipv6No && ipDemandBaseline.NetworkType == constant.IpDemandNetworkTypeIpv6 {
						// ipv4交付，跳过vlan id为ipv6的
						continue
					}
					ipDemandPlannings = append(ipDemandPlannings, &entity.IpDemandPlanning{
						PlanId:          planId,
						LogicalGrouping: logicGroup.LogicalGrouping,
						SegmentType:     ipDemandBaseline.Explain,
						NetworkType:     ipDemandBaseline.NetworkType,
						Vlan:            ipDemandBaseline.Vlan,
						CNum:            ipDemandBaseline.AssignNum,
						Describe:        ipDemandBaseline.Description,
						AddressPlanning: ipDemandBaseline.IpSuggestion,
						CreateTime:      now,
						UpdateTime:      now,
					})
				}
			}
		}
		// ip需求表保存
		if len(ipDemands) > 0 {
			err = ip_demand.DeleteIpDemandPlanningByPlanId(tx, planId)
			if err != nil {
				return err
			}
		}
		if err = ip_demand.SaveBatch(tx, ipDemandPlannings); err != nil {
			return err
		}
		// 软件bom计算并保存
		if err = software_bom.SaveSoftwareBomPlanning(tx, planId); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Errorf("[SaveDeviceList] save device list error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, true)
	return
}

func NetworkDeviceListDownload(context *gin.Context) {
	param := context.Param("planId")
	planId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[IpDemandListDownload] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	fileName, exportResponseDataList, err := ExportNetworkDeviceListByPlanId(planId)
	if err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	total := 0
	for _, response := range exportResponseDataList {
		nums, _ := strconv.Atoi(response.Num)
		total += nums
	}
	// 手动添加合计行
	lastData := NetworkDeviceListExportResponse{
		Num: "总计:" + strconv.Itoa(total) + "台",
	}
	exportResponseDataList = append(exportResponseDataList, lastData)
	_ = excel.NormalDownLoad(fileName, "网络设备清单", "", false, exportResponseDataList, context.Writer)
	return
}

func checkRequest(request *Request) error {
	if request.PlanId == 0 {
		return errors.New("方案ID参数为空")
	}
	if request.AwsServerNum == 0 {
		return errors.New("ASW下连服务器数参数为空")
	}
	if request.AwsBoxNum == 0 {
		return errors.New("每组ASW个数为空")
	}
	if request.NetworkModel == 0 {
		return errors.New("组网模型参数为空")
	}
	if util.IsBlank(request.Brand) {
		return errors.New("厂商参数为空")
	}
	if request.VersionId == 0 {
		return errors.New("版本ID参数为空")
	}
	return nil
}

// 匹配组网模型处理网络设备清单数据
func transformNetworkDeviceList(versionId int64, request *Request, roleBaseLine []entity.NetworkDeviceRoleBaseline, nodeRoleServerNumMap map[int64]int) ([]NetworkDevices, error) {
	var response []NetworkDevices
	networkModel := request.NetworkModel
	/**
	1.循环该数据匹配组网模型， 取出相应的字段 获取到节点角色 无则continue
	2.服务器规划数据满足节点角色的数据  数量累加 得到服务器数量
	3.服务器数量小于asw下连服务器 则直接逻辑分组为最小单元数
	4.服务器数量大于asw下连服务器数量，相除整除直接取商、有余数则商+1为单元数
	5.设备数量= 单元数*单元设备数量
	6.查询网络设备基线数据 根据版本号、网络设备角色、网络版本、厂商
	7.构建网络设备清单列表 两层循环 外层是单元数循环、内层是单元设备数量循环
	*/
	model := constant.NetworkModelYes
	aswNum := make(map[int64]int)
	// TODO roleBaseLine把OASW这条数据移到最后处理
	for _, deviceRole := range roleBaseLine {
		if constant.SeparationOfTwoNetworks == networkModel {
			// 两网分离
			model = deviceRole.TwoNetworkIso
		} else if constant.TripleNetworkSeparation == networkModel {
			// 三网分离
			model = deviceRole.ThreeNetworkIso
		} else {
			// 三网合一
			model = deviceRole.TriplePlay
		}
		networkDevices, err := dealNetworkModel(versionId, request, model, deviceRole, nodeRoleServerNumMap, aswNum)
		if err != nil {
			return nil, err
		}
		if networkDevices != nil && len(networkDevices) > 0 {
			response = append(response, networkDevices...)
		}
	}
	return response, nil
}

// 根据网络设备角色基线计算数据
func dealNetworkModel(versionId int64, request *Request, networkModel int, roleBaseLine entity.NetworkDeviceRoleBaseline, nodeRoleServerNumMap map[int64]int, aswNum map[int64]int) ([]NetworkDevices, error) {
	if constant.NetworkModelNo == networkModel {
		return nil, nil
	}
	funcCompoName := roleBaseLine.FuncCompoName
	funcCompoCode := roleBaseLine.FuncCompoCode
	networkRoleId := roleBaseLine.Id
	brand := request.Brand
	awsServerNum := request.AwsServerNum
	deviceType := request.DeviceType
	var response []NetworkDevices
	deviceModels, _ := GetModelsByVersionIdAndRoleAndBrand(versionId, networkRoleId, brand, deviceType)
	var deviceModel string
	var confOverview string
	var bomId string
	if len(deviceModels) == 0 {
		log.Errorf("[getModelsByVersionIdAndRoleAndBrandAndNetworkConfig] 获取网络设备型号为空")
	} else {
		deviceModel = deviceModels[0].DeviceModel
		confOverview = deviceModels[0].ConfOverview
		bomId = deviceModels[0].BomId
	}
	if constant.NeedQueryOtherTable == networkModel {
		var serverNum = 0
		var nodeRoleOrNetworkRoleIds []int64
		/**
		1.根据网络设备角色ID和组网模型查询关联表获取节点角色ID
		2.为空则return
		3.取其中一条判断是节点角色还是网络设备角色（这里默认每个网络设备角色的关联类型要么全是节点角色，要么全是网络设备角色）
		4.循环关联表数据向nodeRoles添加数据 要注意数量
		*/
		modelRoleRels, err := SearchModelRoleRelByRoleIdAndNetworkModel(networkRoleId, request.NetworkModel)
		if err != nil {
			return nil, err
		}
		if len(modelRoleRels) == 0 {
			log.Errorf("[searchModelRoleRelByRoleIdAndNetworkModel] 获取节点角色关联表数据为空, %v", err)
			return nil, nil
		}
		associatedType := modelRoleRels[0].AssociatedType
		for _, roleRel := range modelRoleRels {
			for i := 1; i <= roleRel.RoleNum; i++ {
				nodeRoleOrNetworkRoleIds = append(nodeRoleOrNetworkRoleIds, roleRel.RoleId)
			}
		}
		log.Infof("角色编码=%v,查询到的节点角色或者设备角色ID=%v", funcCompoCode, nodeRoleOrNetworkRoleIds)
		for _, nodeRoleOrNetworkRoleId := range nodeRoleOrNetworkRoleIds {
			if constant.NodeRoleType == associatedType {
				serverNum += nodeRoleServerNumMap[nodeRoleOrNetworkRoleId]
			} else {
				serverNum += aswNum[nodeRoleOrNetworkRoleId]
			}
		}
		if serverNum == 0 {
			return nil, nil
		}
		log.Infof("角色编码=%v,统计出来的服务器数量=%v", funcCompoCode, serverNum)
		var accessSwitchNum int
		if constant.NodeRoleType == associatedType {
			accessSwitchNum = int(math.Ceil(float64(serverNum) / float64(awsServerNum)))
		} else {
			// 如果是当前设备是连接网络设备的交换机，则是1对1，OASW
			accessSwitchNum = serverNum
		}
		aswNum[networkRoleId] = accessSwitchNum
		response, _ = buildDto(accessSwitchNum, roleBaseLine.UnitDeviceNum, funcCompoName, funcCompoCode, brand, deviceModel, deviceModels, response, networkRoleId, confOverview, bomId)
	} else if constant.NetworkModelYes == networkModel {
		// 固定数量计算
		response, _ = buildDto(roleBaseLine.MinimumNumUnit, roleBaseLine.UnitDeviceNum, funcCompoName, funcCompoCode, brand, deviceModel, deviceModels, response, networkRoleId, confOverview, bomId)
	}
	return response, nil
}

// 组装设备清单
func buildDto(groupNum int, deviceNum int, funcCompoName string, funcCompoCode string, brand string, deviceModel string, deviceModels []NetworkDeviceModel, response []NetworkDevices, deviceRoleId int64, confOverview string, bomId string) ([]NetworkDevices, error) {
	for i := 1; i <= groupNum; i++ {
		logicalGrouping := funcCompoCode + "-" + strconv.Itoa(i)
		for j := 1; j <= deviceNum; j++ {
			deviceId := logicalGrouping + "." + strconv.Itoa(j)
			networkDevice := NetworkDevices{
				NetworkDeviceRole:     funcCompoCode,
				NetworkDeviceRoleName: funcCompoName,
				NetworkDeviceRoleId:   deviceRoleId,
				LogicalGrouping:       logicalGrouping,
				DeviceId:              deviceId,
				Brand:                 brand,
				DeviceModel:           deviceModel,
				DeviceModels:          deviceModels,
				ConfOverview:          confOverview,
				BomId:                 bomId,
			}
			response = append(response, networkDevice)
		}
	}
	return response, nil
}

func queryServerPlanningListByPlanId(planId int64) ([]entity.ServerPlanning, error) {
	var serverPlanningList []entity.ServerPlanning
	if err := data.DB.Where("plan_id = ? AND delete_state = 0", planId).Find(&serverPlanningList).Error; err != nil {
		return serverPlanningList, err
	}
	return serverPlanningList, nil
}

func ListNetworkShelve(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Error(err)
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	networkShelveList, err := GetNetworkShelveList(request.PlanId)
	if err != nil {
		log.Errorf("ListNetworkShelve error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, networkShelveList)
	return
}

func DownloadNetworkShelveTemplate(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	response, fileName, err := GetDownloadNetworkShelveTemplate(planId)
	if err != nil {
		log.Errorf("DownloadNetworkShelveTemplate error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if err = excel.NormalDownLoad(fileName, "网络设备上架模板", "", false, response, c.Writer); err != nil {
		log.Errorf("下载错误：", err)
	}
	return
}

func UploadShelve(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	// 上传文件处理
	file, err := c.FormFile("file")
	if err != nil {
		log.Error(err)
		result.Failure(c, "文件错误", http.StatusBadRequest)
		return
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "networkShelve", time.Now().Unix(), rand.Uint32())
	if err = c.SaveUploadedFile(file, filePath); err != nil {
		log.Error(err)
		result.Failure(c, "保存文件错误", http.StatusInternalServerError)
		return
	}
	f, err := excelize.OpenFile(filePath)
	defer func() {
		if err = f.Close(); err != nil {
			log.Errorf("excelize close error: %v", err)
		}
		if err = os.Remove(filePath); err != nil {
			log.Errorf("os removeFile error: %v", err)
		}
	}()
	if err != nil {
		log.Error(err)
		result.Failure(c, "打开文件错误", http.StatusInternalServerError)
		return
	}
	var networkDeviceShelveDownload []NetworkDeviceShelveDownload
	if err = excel.ImportBySheet(f, &networkDeviceShelveDownload, "网络设备上架模板", 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(c, "解析文件错误", http.StatusInternalServerError)
		return
	}
	userId := user.GetUserId(c)
	if err = UploadNetworkShelve(planId, networkDeviceShelveDownload, userId); err != nil {
		log.Errorf("UploadNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func SaveShelve(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error(err)
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	request.UserId = user.GetUserId(c)
	if err := SaveNetworkShelve(request); err != nil {
		log.Errorf("SaveNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func DownloadNetworkShelve(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId不能为空", http.StatusBadRequest)
		return
	}
	response, fileName, err := GetDownloadNetworkShelve(planId)
	if err != nil {
		log.Errorf("DownloadNetworkShelve error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if err = excel.NormalDownLoad(fileName, "网络设备上架清单", "", false, response, c.Writer); err != nil {
		log.Errorf("下载错误：", err)
	}
	return
}
