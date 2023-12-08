package network_device

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"code.cestc.cn/ccos/cnm/ops-base/utils"
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
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/baseline"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/ip_demand"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/plan"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/server"
)

// func GetCountBoxNum(c *gin.Context) {
//	request := &Request{}
//	if err := c.ShouldBindQuery(&request); err != nil {
//		log.Errorf("get device plan bind param error: ", err)
//		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
//		return
//	}
//	if err := checkRequest(request); err != nil {
//		result.Failure(c, err.Error(), http.StatusBadRequest)
//		return
//	}
//	var boxCount BoxTotalResponse
//	boxCount.Count = 8
//	awsServerNum := request.AwsServerNum
//	planId := request.PlanId
//	//TODO 根据方案id查询服务器规划表
//
//	serverNumMap := make(map[string]int, 3)
//	serverNumMap[MASW] = 0
//	serverNumMap[VASW] = 0
//	serverNumMap[StorSASW] = 0
//	//TODO 计算机柜数量
//	result.Success(c, boxCount)
//	return
// }

func GetDevicePlanByPlanId(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	// 根据方案ID查询网络设备规划
	devicePlan, err := searchDevicePlanByPlanId(planId)
	if err != nil {
		log.Errorf("[searchDevicePlanByPlanId] search device plan by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if devicePlan.PlanId == 0 {
		devicePlan = nil
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
	cloudPlatformType := request.CloudPlatformType
	baselineVersion := request.BaselineVersion
	if len(cloudPlatformType) == 0 || len(baselineVersion) == 0 || planId == 0 {
		result.FailureWithMsg(c, errorcodes.InvalidParam, http.StatusBadRequest, errorcodes.ParamError)
		return
	}
	// 根据云产品版本和云平台类型查询版本ID
	versionId, err := getVersionId(baselineVersion, cloudPlatformType)
	if err != nil {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, err.Error())
		return
	}
	// 根据方案id查询云产品规划信息  取其中一条拿服务器基线表ID
	serverPlanningList, err := server.QueryServerPlanningListByPlanId(planId)
	if err != nil {
		log.Errorf("[QueryServerPlanningListByPlanId] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(serverPlanningList) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.ServerPlanningListEmpty)
		return
	}
	serverBaselineId := serverPlanningList[0].ServerBaselineId
	// 根据服务器基线ID查询服务器基线表 获取网络版本
	serverBaseline, err := baseline.QueryServiceBaselineById(serverBaselineId)
	if err != nil {
		log.Errorf("[QueryServiceBaselineById] search baseline by id error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if serverBaseline.Id == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.ServerBaselineEmpty)
		return
	}
	networkVersion := serverBaseline.NetworkInterface
	// 根据服务版本id和网络版本查询网络设备基线表查厂商  去重
	brands, err := getBrandsByVersionIdAndNetworkVersion(versionId, networkVersion)
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
	var response []NetworkDevices
	// 根据方案ID查询网络设备规划表 没有则保存，有则更新
	planId := request.PlanId
	devicePlan, err := searchDevicePlanByPlanId(planId)
	if err != nil {
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	// 根据云产品版本和云平台类型查询版本ID
	versionId, err := getVersionId(request.BaselineVersion, request.CloudPlatformType)
	if err != nil {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, err.Error())
		return
	}
	// 根据方案id查询服务器规划
	serverPlanningList, err := server.QueryServerPlanningListByPlanId(planId)
	if err != nil {
		log.Errorf("[QueryServerPlanningListByPlanId] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(serverPlanningList) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.ServerPlanningListEmpty)
		return
	}
	// 服务器规划数据转为map
	var nodeRoleServerNumMap = make(map[int64]int)
	for _, value := range serverPlanningList {
		nodeRoleServerNumMap[value.NodeRoleId] = value.Number
	}
	// 根据服务器基线id查询服务器基线表获取网络接口
	serviceBaselineId := serverPlanningList[0].ServerBaselineId
	serverBaseline, err := baseline.QueryServiceBaselineById(serviceBaselineId)
	if err != nil {
		log.Errorf("[QueryServiceBaselineById] search baseline by id error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	networkInterface := serverBaseline.NetworkInterface
	// 根据版本号查询出网络设备角色基线数据
	deviceRoleBaseline, err := searchDeviceRoleBaselineByVersionId(versionId)
	if err != nil {
		log.Errorf("[searchDeviceRoleBaselineByVersionId] search device role baseline error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(deviceRoleBaseline) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.NetworkDeviceRoleBaselineEmpty)
		return
	}
	response, err = transformNetworkDeviceList(versionId, networkInterface, request, deviceRoleBaseline, nodeRoleServerNumMap)
	if err != nil {
		log.Errorf("[transformNetworkDeviceList] error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if devicePlan.Id == 0 {
		err = createDevicePlan(request)
	} else {
		err = updateDevicePlan(request, *devicePlan)
	}
	if err != nil {
		log.Errorf("[saveOrUpdateDevicePlan] save or update device plan error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, response)
	return
}

func SaveDeviceList(c *gin.Context) {
	req := &Request{}
	var networkDeviceList []*entity.NetworkDeviceList
	var ipDemandPlannings []*entity.IPDemandPlanning
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
		device := new(entity.NetworkDeviceList)
		device.PlanId = planId
		device.NetworkDeviceRole = networkDevice.NetworkDeviceRole
		device.NetworkDeviceRoleId = networkDevice.NetworkDeviceRoleId
		device.NetworkDeviceRoleName = networkDevice.NetworkDeviceRoleName
		device.ConfOverview = networkDevice.ConfOverview
		device.LogicalGrouping = networkDevice.LogicalGrouping
		device.DeviceId = networkDevice.DeviceId
		device.Brand = networkDevice.Brand
		device.DeviceModel = networkDevice.DeviceModel
		device.CreateTime = now
		device.UpdateTime = now
		device.DeleteState = 0
		networkDeviceList = append(networkDeviceList, device)
	}
	deviceList, err := searchDeviceListByPlanId(planId)
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
	versionId, err := getVersionId(req.BaselineVersion, req.CloudPlatformType)
	if err != nil {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, err.Error())
		return
	}
	// 根据版本ID查询ip需求基线数据
	ipDemandBaselines, err := ip_demand.GetIpDemandBaselineByVersionId(versionId)
	if err != nil {
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(ipDemandBaselines) == 0 {
		result.FailureWithMsg(c, errorcodes.SystemError, http.StatusInternalServerError, errorcodes.IpDemandBaselineEmpty)
	}
	// 转ip基线ID Map
	ipBaselineIdMap := util.ListToMaps(ipDemandBaselines, "ID")
	log.Infof("IP需求基线表转为map后的数据=%v", ipBaselineIdMap)
	//userId := user.GetUserId(c)
	userId := "1"
	err = data.DB.Transaction(func(tx *gorm.DB) error {
		if len(deviceList) > 0 {
			// 失效库里保存的
			err = expireDeviceListByPlanId(tx, planId)
			if err != nil {
				return err
			}
		}
		// 更新方案表的状态
		err = plan.UpdatePlanStage(tx, planId, constant.PLANNED, userId, constant.BUSINESS_END)
		if err != nil {
			return err
		}
		// 批量保存网络设备清单
		err = SaveBatch(tx, networkDeviceList)
		if err != nil {
			return err
		}
		// 根据方案ID查询网络设备清单
		deviceRoleNum, err := getDeviceRoleGroupNumByPlanId(tx, planId)
		// 转设备角色id和分组数 map
		deviceRoleIdMap := util.ListToMap(deviceRoleNum, "DeviceRoleId")
		log.Infof("deviceRoleIdMap=%v", deviceRoleIdMap)
		for _, value := range ipBaselineIdMap {
			var num = 0
			dto := new(ip_demand.IpDemandBaselineDto)
			var needCount bool
			// 根据IP需求规划表的关联设备组进行 网络设备清单分组累加
			for i, val := range value {
				v := val.(*ip_demand.IpDemandBaselineDto)
				log.Infof("IpDemandBaselineDto=%v", v)
				if i == 0 {
					dto = v
				}
				deviceRoleId := v.DeviceRoleId
				groupNum, ok := deviceRoleIdMap[strconv.FormatInt(deviceRoleId, 10)]
				if ok {
					needCount = true
					deviceGroupNum := groupNum.(*DeviceRoleGroupNum)
					num += deviceGroupNum.GroupNum
				}
			}
			if needCount {
				ipDemandPlanning := new(entity.IPDemandPlanning)
				ipDemandPlanning.PlanId = planId
				ipDemandPlanning.SegmentType = dto.Explain
				ipDemandPlanning.Vlan = dto.Vlan
				assignNum, _ := utils.String2Float(dto.AssignNum)
				cNum := assignNum * float64(num)
				ipDemandPlanning.Cnum = fmt.Sprintf("%f", cNum)
				ipDemandPlanning.Describe = dto.Description
				ipDemandPlanning.AddressPlanning = dto.IpSuggestion
				ipDemandPlanning.CreateTime = now
				ipDemandPlanning.UpdateTime = now
				ipDemandPlannings = append(ipDemandPlannings, ipDemandPlanning)
			}
		}
		// ip需求表保存
		if len(ipDemands) > 0 {
			err = ip_demand.DeleteIpDemandPlanningByPlanId(tx, planId)
			if err != nil {
				return err
			}
		}
		err = ip_demand.SaveBatch(tx, ipDemandPlannings)
		return err
	})
	if err != nil {
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
	fileName, exportResponseDataList, err := exportNetworkDeviceListByPlanId(planId)
	if err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
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
	if util.IsBlank(request.CloudPlatformType) {
		return errors.New("云平台类型参数为空")
	}
	if util.IsBlank(request.BaselineVersion) {
		return errors.New("云产品版本参数为空")
	}
	return nil
}

// 匹配组网模型处理网络设备清单数据
func transformNetworkDeviceList(versionId int64, networkInterface string, request *Request, roleBaseLine []entity.NetworkDeviceRoleBaseline, nodeRoleServerNumMap map[int64]int) ([]NetworkDevices, error) {
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
		if constant.SEPARATION_OF_TWO_NETWORKS == networkModel {
			// 两网分离
			model = deviceRole.TwoNetworkIso
		} else if constant.TRIPLE_NETWORK_SEPARATION == networkModel {
			// 三网分离
			model = deviceRole.ThreeNetworkIso
		} else {
			// 三网合一
			model = deviceRole.TriplePlay
		}
		networkDevices, err := dealNetworkModel(versionId, networkInterface, request, model, deviceRole, nodeRoleServerNumMap, aswNum)
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
func dealNetworkModel(versionId int64, networkInterface string, request *Request, networkModel int, roleBaseLine entity.NetworkDeviceRoleBaseline, nodeRoleServerNumMap map[int64]int, aswNum map[int64]int) ([]NetworkDevices, error) {
	if constant.NetworkModelNo == networkModel {
		return nil, nil
	}
	funcCompoName := roleBaseLine.FuncCompoName
	funcCompoCode := roleBaseLine.FuncCompoCode
	id := roleBaseLine.Id
	brand := request.Brand
	awsServerNum := request.AwsServerNum
	var response []NetworkDevices
	deviceModels, _ := getModelsByVersionIdAndRoleAndBrandAndNetworkConfig(versionId, networkInterface, id, brand)
	if len(deviceModels) == 0 {
		log.Errorf("[getModelsByVersionIdAndRoleAndBrandAndNetworkConfig] 获取网络设备型号为空")
		return nil, nil
	}
	deviceModel := deviceModels[0].DeviceModel
	confOverview := deviceModels[0].ConfOverview
	//TODO 根据设备类型获取唯一型号
	if len(deviceModels) > 1 {

	}
	if constant.NeedQueryOtherTable == networkModel {
		var serverNum = 0
		var nodeRoles []int64
		/**
		1.根据网络设备角色ID和组网模型查询关联表获取节点角色ID
		2.为空则return
		3.取其中一条判断是节点角色还是网络设备角色（这里默认每个网络设备角色的关联类型要么全是节点角色，要么全是网络设备角色）
		4.循环关联表数据向nodeRoles添加数据 要注意数量
		*/
		modelRoleRels, err := searchModelRoleRelByRoleIdAndNetworkModel(id, request.NetworkModel)
		if err != nil {
			return nil, err
		}
		if len(modelRoleRels) == 0 {
			log.Errorf("[searchModelRoleRelByRoleIdAndNetworkModel] 获取节点角色关联表数据为空, %v", err)
			return nil, nil
		}
		associatedType := modelRoleRels[0].AssociatedType
		for _, roleRel := range modelRoleRels {
			roleNum := roleRel.RoleNum
			roleId := roleRel.RoleId
			for i := 1; i <= roleNum; i++ {
				nodeRoles = append(nodeRoles, roleId)
			}
		}
		log.Infof("角色编码=%v,查询到的节点角色或者设备角色ID=%v", funcCompoCode, nodeRoles)
		for _, nodeRoleId := range nodeRoles {
			if constant.NodeRoleType == associatedType {
				serverNum += nodeRoleServerNumMap[nodeRoleId]
			} else {
				serverNum += aswNum[nodeRoleId]
			}
		}
		if constant.NodeRoleType == associatedType {
			aswNum[id] = serverNum
		}
		log.Infof("角色编码=%v,统计出来的服务器数量=%v", funcCompoCode, serverNum)
		var minimumNumUnit = 1
		if serverNum > awsServerNum {
			discuss := serverNum / awsServerNum
			remainder := serverNum % awsServerNum
			if remainder == 0 {
				minimumNumUnit = discuss
			} else {
				minimumNumUnit += discuss
			}
		}
		response, _ = buildDto(minimumNumUnit, roleBaseLine.UnitDeviceNum, funcCompoName, funcCompoCode, brand, deviceModel, deviceModels, response, id, confOverview)
	} else if constant.NetworkModelYes == networkModel {
		// 固定数量计算
		response, _ = buildDto(roleBaseLine.MinimumNumUnit, roleBaseLine.UnitDeviceNum, funcCompoName, funcCompoCode, brand, deviceModel, deviceModels, response, id, confOverview)
	}
	return response, nil
}

// 组装设备清单
func buildDto(groupNum int, deviceNum int, funcCompoName string, funcCompoCode string, brand string, deviceModel string, deviceModels []NetworkDeviceModel, response []NetworkDevices, deviceRoleId int64, confOverview string) ([]NetworkDevices, error) {
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
			}
			response = append(response, networkDevice)
		}
	}
	return response, nil
}

func getVersionId(baselineVersion string, cloudPlatformType string) (int64, error) {
	// 根据方案id查询版本id
	softwareVersion, err := querySoftwareVersionByVersion(baselineVersion, cloudPlatformType)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error("获取云产品版本异常, error %v", err)
		return 0, err
	}
	versionId := softwareVersion.Id
	if versionId == 0 {
		return 0, errors.New(errorcodes.NotFoundVersionMsg)
	}
	return versionId, nil
}
