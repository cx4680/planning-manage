package ip_demand

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"
	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/data"
	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/excel"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/util"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/baseline"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/cloud_product"
)

func IpDemandListDownload(context *gin.Context) {
	param := context.Param("planId")
	planId, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Errorf("[IpDemandListDownload] invalid param error, %v", err)
		result.Failure(context, errorcodes.InvalidParam, http.StatusBadRequest)
		return
	}
	fileName, exportResponseDataList, err := ExportIpDemandPlanningByPlanId(planId)
	if err != nil {
		log.Errorf("[exportIpDemandPlanningByPlanId] error, %v", err)
		result.Failure(context, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	_ = excel.NormalDownLoad(fileName, "IP需求清单", "", false, exportResponseDataList, context.Writer)
	return
}

func List(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Errorf("getIpDemandList bind param error: ", err)
	}
	if request.PlanId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	list, err := GetIpDemandPlanningList(request.PlanId)
	if err != nil {
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, list)
	return
}

func Upload(c *gin.Context) {
	planId, err := strconv.ParseInt(c.Param("planId"), 10, 64)
	if planId == 0 {
		result.Failure(c, "planId参数为空", http.StatusBadRequest)
		return
	}
	// 上传文件处理
	file, err := c.FormFile("file")
	if err != nil {
		log.Error(err)
		result.Failure(c, "文件错误", http.StatusBadRequest)
		return
	}
	filePath := fmt.Sprintf("%s/%s-%d-%d.xlsx", "exampledir", "ipDemand", time.Now().Unix(), rand.Uint32())
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
	var ipDemandPlanningExportResponse []IpDemandPlanningExportResponse
	if err = excel.ImportBySheet(f, &ipDemandPlanningExportResponse, "IP需求清单", 0, 1); err != nil {
		log.Errorf("excel import error: %v", err)
		result.Failure(c, "解析文件错误", http.StatusInternalServerError)
		return
	}
	if err = UploadIpDemand(planId, ipDemandPlanningExportResponse); err != nil {
		log.Errorf("UploadNetworkShelve error, %v", err)
		result.Failure(c, err.Error(), http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func Save(c *gin.Context) {
	request := &Request{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error(err)
	}
	planId := request.PlanId
	err := data.DB.Transaction(func(tx *gorm.DB) error {
		// 更新方案表的状态
		if err := SaveIpDemand(tx, planId); err != nil {
			return err
		}
		cloudProductPlannings, err := cloud_product.ListCloudProductPlanningByPlanId(planId)
		if err != nil {
			return err
		}
		versionId := cloudProductPlannings[0].VersionId
		networkDeviceRoleBaselines, err := baseline.QueryNetworkDeviceRoleBaselineByVersionId(versionId)
		if err != nil {
			return err
		}
		var accessNetworkDeviceRoleBaselines []entity.NetworkDeviceRoleBaseline
		for _, networkDeviceRoleBaseline := range networkDeviceRoleBaselines {
			if strings.Contains(networkDeviceRoleBaseline.FuncType, constant.AccessNetworkDeviceRoleKeyword) {
				accessNetworkDeviceRoleBaselines = append(accessNetworkDeviceRoleBaselines, networkDeviceRoleBaseline)
			}
		}
		networkShelves, err := GetNetworkShelveList(planId)
		if err != nil {
			return err
		}
		// 获取全局配置的规划文件sheet1的网络设备列表，只统计接入交换机，并且每个逻辑分组只用统计第一个设备
		var accessNetworkShelves []entity.NetworkDeviceShelve
		accessNetworkDeviceLoginGroupMap := make(map[string]int64)
		for _, networkShelf := range networkShelves {
			for _, networkDeviceRoleBaseline := range accessNetworkDeviceRoleBaselines {
				if strings.Contains(networkShelf.DeviceLogicalId, networkDeviceRoleBaseline.FuncType) {
					if _, ok := accessNetworkDeviceLoginGroupMap[networkShelf.DeviceLogicalId]; !ok {
						accessNetworkDeviceLoginGroupMap[networkShelf.DeviceLogicalId] = networkShelf.Id
						accessNetworkShelves = append(accessNetworkShelves, *networkShelf)
					}
				}
			}
		}
		ipDemandPlannings, err := SearchIpDemandPlanningByPlanId(planId)
		if err != nil {
			return err
		}
		ipDemandPlanningMap := make(map[string]string)
		for _, ipDemandPlanning := range ipDemandPlannings {
			if ipDemandPlanning.Address != "" {
				key := fmt.Sprintf("%s_%d_%s", ipDemandPlanning.LogicalGrouping, ipDemandPlanning.NetworkType, ipDemandPlanning.Vlan)
				ipDemandPlanningMap[key] = ipDemandPlanning.Address
			}
		}
		var networkDeviceIps []entity.NetworkDeviceIp
		for _, accessNetworkShelf := range accessNetworkShelves {
			networkDeviceIp := entity.NetworkDeviceIp{
				PlanId:          planId,
				LogicalGrouping: accessNetworkShelf.DeviceLogicalId,
			}
			err = ConvertNetworkDeviceIp(accessNetworkShelf, ipDemandPlanningMap, &networkDeviceIp)
			if err != nil {
				return err
			}
			networkDeviceIps = append(networkDeviceIps, networkDeviceIp)
		}
		if err = CreateNetworkDeviceIp(tx, networkDeviceIps); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Errorf("[SaveNetworkShelve] save network shelve error, %v", err)
		result.Failure(c, errorcodes.SystemError, http.StatusInternalServerError)
		return
	}
	result.Success(c, nil)
	return
}

func ConvertNetworkDeviceIp(accessNetworkShelf entity.NetworkDeviceShelve, ipDemandPlanningMap map[string]string, networkDeviceIp *entity.NetworkDeviceIp) error {
	pxeIpv4Key := fmt.Sprintf("%s_%d_%s", accessNetworkShelf.DeviceLogicalId, constant.IpDemandNetworkTypeIpv4, constant.NetworkDevicePxeVlanId)
	address, ok := ipDemandPlanningMap[pxeIpv4Key]
	if ok {
		networkDeviceIp.PxeSubnet = address
		ips, err := util.ParseCIDR(address)
		if err != nil {
			return err
		}
		if len(ips) < 59 {
			return fmt.Errorf("pxe ip数量小于59")
		}
		networkDeviceIp.PxeSubnetRange = strings.Join(ips[:59], constant.Comma)
		networkDeviceIp.PxeNetworkGateway = ips[len(ips)-2]
	}
	manageIpv4Key := fmt.Sprintf("%s_%d_%s", accessNetworkShelf.DeviceLogicalId, constant.IpDemandNetworkTypeIpv4, constant.NetworkDeviceManageVlanId)
	address, ok = ipDemandPlanningMap[manageIpv4Key]
	if ok {
		networkDeviceIp.ManageSubnet = address
		ips, err := util.ParseCIDR(address)
		if err != nil {
			return err
		}
		if len(ips) < 3 {
			return fmt.Errorf("管理网子网ip数量小于3")
		}
		networkDeviceIp.ManageNetworkGateway = ips[len(ips)-2]
	}
	manageIpv6Key := fmt.Sprintf("%s_%d_%s", accessNetworkShelf.DeviceLogicalId, constant.IpDemandNetworkTypeIpv6, constant.NetworkDeviceManageVlanId)
	address, ok = ipDemandPlanningMap[manageIpv6Key]
	if ok {
		networkDeviceIp.ManageIpv6Subnet = address
		addressSplits := strings.Split(address, constant.ForwardSlash)
		networkDeviceIp.ManageIpv6NetworkGateway = addressSplits[0] + "1"
	}
	bizIpv4Key := fmt.Sprintf("%s_%d_%s", accessNetworkShelf.DeviceLogicalId, constant.IpDemandNetworkTypeIpv4, constant.NetworkDeviceBizVlanId)
	address, ok = ipDemandPlanningMap[bizIpv4Key]
	if ok {
		networkDeviceIp.BizSubnet = address
		ips, err := util.ParseCIDR(address)
		if err != nil {
			return err
		}
		if len(ips) < 3 {
			return fmt.Errorf("业务网子网ip数量小于3")
		}
		networkDeviceIp.BizNetworkGateway = ips[len(ips)-2]
	}
	storageIpv4Key := fmt.Sprintf("%s_%d_%s", accessNetworkShelf.DeviceLogicalId, constant.IpDemandNetworkTypeIpv4, constant.NetworkDeviceStorageVlanId)
	address, ok = ipDemandPlanningMap[storageIpv4Key]
	if ok {
		networkDeviceIp.StorageFrontNetwork = address
		ips, err := util.ParseCIDR(address)
		if err != nil {
			return err
		}
		if len(ips) < 3 {
			return fmt.Errorf("存储前端网ip数量小于3")
		}
		networkDeviceIp.StorageFrontNetworkGateway = ips[len(ips)-2]
	}
	return nil
}
