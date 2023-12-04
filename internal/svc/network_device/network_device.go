package network_device

import (
	"code.cestc.cn/zhangzhi/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/result"
	"code.cestc.cn/zhangzhi/planning-manage/internal/pkg/util"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"net/http"
	"strconv"
)

type Request struct {
	PlanId                int64  `form:"planId"`
	Brand                 string `form:"brand"`
	ApplicationDispersion string `form:"applicationDispersion"`
	AwsServerNum          int    `form:"awsServerNum"`
	AwsBoxNum             int    `form:"awsBoxNum"`
	TotalBoxNum           int    `form:"totalBoxNum"`
	Ipv6                  string `form:"ipv6"`
	NetworkModel          string `form:"networkModel"`
	OpenDpdk              string `form:"openDpdk"`
}

type NetworkDevices struct {
	PlanId            int64                `form:"planId"`
	NetworkDeviceRole string               `form:"networkDeviceRole"`
	LogicalGrouping   string               `form:"logicalGrouping"`
	DeviceId          string               `form:"deviceId"`
	Brand             string               `form:"brand"`
	DeviceModel       string               `form:"deviceModel"`
	DeviceModels      []NetworkDeviceModel `form:"deviceModels"`
}

type NetworkDeviceModel struct {
	ConfigurationOverview string `form:"configurationOverview"`
	DeviceModel           string `form:"deviceModel"`
}

//const (
//	MASW = "MASW"
//	VASW = "VASW"
//	StorSASW = "StorSASW"
//)

//func GetCountBoxNum(c *gin.Context) {
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
//}

func GetDevicePlanByPlanId(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	//根据方案ID查询网络设备规划
	devicePlan, err := searchDevicePlanByPlanId(planId)
	if err != nil {
		log.Errorf("[searchDevicePlanByPlanId] search device plan by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, devicePlan)
	return
}

func GetBrandsByPlanId(c *gin.Context) {
	planId, _ := strconv.ParseInt(c.Param("planId"), 10, 64)
	//TODO 根据方案id查询版本id
	//var versionId string
	//TODO 根据方案id查询云产品规划信息  取其中一条拿服务器基线表ID

	//TODO 根据服务器基线ID查询服务器基线表 获取网络版本
	//var networkVersion string
	//TODO 根据服务版本id和网络版本查询网络设备基线表查厂商  去重

	result.Success(c, planId)
	return
}

func ListNetworkDevices(c *gin.Context) {
	request := Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("list network devices bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if err := checkRequest(request); err != nil {
		result.Failure(c, err.Error(), http.StatusBadRequest)
		return
	}
	var response []NetworkDevices
	//根据方案ID查询网络设备规划表 没有则保存，有则更新
	planId := request.PlanId
	//awsServerNum := request.AwsServerNum
	//brand := request.Brand
	//networkModel := request.NetworkModel
	devicePlan, err := searchDevicePlanByPlanId(planId)
	if err != nil {
		log.Errorf("[searchDevicePlanByPlanId] search device plan by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	//err = data.DB.Transaction(func(tx *gorm.DB) error {
	//	return err
	//})
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
	//TODO 根据方案id查询服务器规划

	//TODO 根据版本号查询出网络设备角色基线数据
	/**
	1.循环该数据 swich case匹配组网模型， 取出相应的字段 获取到节点角色 无则continue
	2.服务器规划数据满足节点角色的数据  数量累加 得到服务器数量
	3.服务器数量小于asw下连服务器 则直接逻辑分组为最小单元数
	4.服务器数量大于asw下连服务器数量，相除整除直接取商、有余数则商+1为单元数
	5.设备数量= 单元数*单元设备数量
	6.查询网络设备基线数据 根据版本号、网络设备角色、网络版本、厂商
	7.构建网络设备清单列表 两层循环 外层是单元数循环、内层是单元设备数量循环
	*/
	result.Success(c, response)
	return
}

func SaveDeviceList(c *gin.Context) {
	var request []NetworkDevices
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("save network devices bind param error: ", err)
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	if len(request) == 0 {
		result.Failure(c, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	planId := request[0].PlanId
	deviceList, err := searchDeviceListByPlanId(planId)
	if err != nil {
		log.Errorf("[searchDeviceListByPlanId] search device list by planId error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	if len(deviceList) > 0 {
		//TODO 删除库里保存的
	}
	//TODO 更新方案表的状态
	//TODO 批量保存网络设备清单
	//TODO ip需求表保存
	result.Success(c, nil)
	return
}

func checkRequest(request Request) error {
	if request.PlanId == 0 {
		return errors.New("方案ID参数为空")
	}
	if request.AwsServerNum == 0 {
		return errors.New("ASW下连服务器数参数为空")
	}
	if request.AwsBoxNum == 0 {
		return errors.New("每组ASW个数为空")
	}
	if util.IsBlank(request.NetworkModel) {
		return errors.New("组网模型参数为空")
	}
	if util.IsBlank(request.Brand) {
		return errors.New("厂商参数为空")
	}
	return nil
}
