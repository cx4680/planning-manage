package http

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/opentrx/seata-golang/v2/pkg/util/log"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/errorcodes"
	"code.cestc.cn/ccos/common/planning-manage/internal/pkg/result"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/cloud_product"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/config_item"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/global_config"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/machine_room"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/server"

	"code.cestc.cn/ccos/common/planning-manage/internal/svc/az"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/cell"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/cloud_platform"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/customer"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/ip_demand"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/network_device"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/plan"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/project"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/region"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/user"

	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/common/planning-manage/internal/svc/baseline"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"

	"code.cestc.cn/ccos/cnm/ops-base/opshttp/middleware"
)

const (
	apiPrefix   = "/api/planning-manage/v1"
	innerPrefix = "/api/planning-manage/v1/inner"
)

func Router(engine *gin.Engine) {

	api := engine.Group(apiPrefix, Auth())
	{
		// user
		userGroup := engine.Group(apiPrefix + "/user")
		{
			// 查询操作列表
			userGroup.POST("/login", middleware.OperatorLog(DefaultEventOpInfo("登录", "login", middleware.OPERATE, middleware.INFO)), user.Login)
			// 新增操作
			userGroup.GET("/logout", middleware.OperatorLog(DefaultEventOpInfo("登出", "logout", middleware.OPERATE, middleware.INFO)), user.Logout)
			// 根据名称查询ldap用户
			userGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("根据关键字查询用户", "listByName", middleware.OPERATE, middleware.INFO)), user.ListByName)
		}

		customerGroup := api.Group("/customer")
		{
			// 分页查询客户列表
			customerGroup.POST("/page", middleware.OperatorLog(DefaultEventOpInfo("分页查询客户列表", "queryCustomerByPage", middleware.LIST, middleware.INFO)), customer.Page)
			// 获取客户列表
			// customerGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("获取客户列表", "listCustomer", middleware.LIST, middleware.INFO)), customer.List)
			// 根据id获取客户
			customerGroup.GET("/:id", middleware.OperatorLog(DefaultEventOpInfo("根据id获取客户", "queryCustomerById", middleware.GET, middleware.INFO)), customer.GetById)
			// 创建客户
			customerGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建客户", "createCustomer", middleware.CREATE, middleware.INFO)), customer.Create)
			// 根据id修改客户
			customerGroup.POST("/update", middleware.OperatorLog(DefaultEventOpInfo("根据id修改客户", "editCustomer", middleware.UPDATE, middleware.INFO)), customer.Update)
			// 根据id删除客户
			// customerGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("根据id删除客户", "deleteCustomer", middleware.DELETE, middleware.INFO)), customer.Delete)
		}

		// 云平台
		cloudPlatformGroup := api.Group("/platform")
		{
			// 查询云平台列表
			cloudPlatformGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询云平台列表", "queryProjectList", middleware.LIST, middleware.INFO)), cloud_platform.List)
			// 创建云平台
			cloudPlatformGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建云平台", "createProject", middleware.CREATE, middleware.INFO)), cloud_platform.Create)
			// 根据id修改云平台
			cloudPlatformGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改云平台", "deleteProject", middleware.UPDATE, middleware.INFO)), cloud_platform.Update)
			// 查询树形图
			cloudPlatformGroup.GET("/tree", middleware.OperatorLog(DefaultEventOpInfo("查询云平台列表", "queryProjectList", middleware.LIST, middleware.INFO)), cloud_platform.Tree)
		}

		// region
		regionGroup := api.Group("/region")
		{
			// 查询region列表
			regionGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询region列表", "queryRegionList", middleware.GET, middleware.INFO)), region.List)
			// 创建region
			regionGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建region", "createRegion", middleware.CREATE, middleware.INFO)), region.Create)
			// 修改region
			regionGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改region", "updateRegion", middleware.UPDATE, middleware.INFO)), region.Update)
			// 删除region
			regionGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("删除方案", "deleteRegion", middleware.DELETE, middleware.INFO)), region.Delete)
		}

		// az
		azGroup := api.Group("/az")
		{
			// 查询az列表
			azGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询az列表", "queryAzList", middleware.LIST, middleware.INFO)), az.List)
			// 创建az
			azGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建az", "createAz", middleware.CREATE, middleware.INFO)), az.Create)
			// 修改az
			azGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改az", "updateAz", middleware.UPDATE, middleware.INFO)), az.Update)
			// 删除az
			azGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("删除az", "deleteAz", middleware.UPDATE, middleware.INFO)), az.Delete)
		}

		// cell
		cellGroup := api.Group("/cell")
		{
			// 查询cell列表
			cellGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询cell列表", "queryCellList", middleware.LIST, middleware.INFO)), cell.List)
			// 创建cell
			cellGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建cell", "createCell", middleware.CREATE, middleware.INFO)), cell.Create)
			// 修改cell
			cellGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改cell", "updateCell", middleware.UPDATE, middleware.INFO)), cell.Update)
			// 删除cell
			cellGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("删除cell", "deleteCell", middleware.UPDATE, middleware.INFO)), cell.Delete)
		}

		// 项目管理
		projectGroup := api.Group("/project")
		{
			// 分页查询项目
			projectGroup.GET("/page", middleware.OperatorLog(DefaultEventOpInfo("分页查询项目", "queryProjectByPage", middleware.LIST, middleware.INFO)), project.Page)
			// 创建项目
			projectGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建项目", "createProject", middleware.CREATE, middleware.INFO)), project.Create)
			// 修改项目
			projectGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改项目", "updateProject", middleware.UPDATE, middleware.INFO)), project.Update)
			// 删除项目
			projectGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("删除项目", "deleteProject", middleware.DELETE, middleware.INFO)), project.Delete)
		}

		// 方案管理
		planGroup := api.Group("/plan")
		{
			// 分页查询方案
			planGroup.GET("/page", middleware.OperatorLog(DefaultEventOpInfo("分页查询方案", "queryPlanByPage", middleware.LIST, middleware.INFO)), plan.Page)
			// 创建方案
			planGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建方案", "createPlan", middleware.CREATE, middleware.INFO)), plan.Create)
			// 修改方案
			planGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改方案", "determinePlanById", middleware.UPDATE, middleware.INFO)), plan.Update)
			// 删除方案
			planGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("删除方案", "deletePlanById", middleware.DELETE, middleware.INFO)), plan.Delete)
		}

		// 服务器规划
		serverGroup := api.Group("/server")
		{
			// 查询服务器规划列表
			serverGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询服务器列表", "queryServerList", middleware.LIST, middleware.INFO)), server.List)
			// 保存服务器规划
			serverGroup.POST("/save", middleware.OperatorLog(DefaultEventOpInfo("创建服务器", "saveServerList", middleware.LIST, middleware.INFO)), server.Save)
			// 查询网络类型列表
			serverGroup.GET("/network/list", middleware.OperatorLog(DefaultEventOpInfo("查询网络类型列表", "queryServerArchList", middleware.LIST, middleware.INFO)), server.NetworkTypeList)
			// 查询cpu类型列表
			serverGroup.GET("/cpu/list", middleware.OperatorLog(DefaultEventOpInfo("查询服务器架构列表", "queryServerArchList", middleware.LIST, middleware.INFO)), server.CpuTypeList)
			// 查询容量规划列表
			serverGroup.GET("/capacity/list", middleware.OperatorLog(DefaultEventOpInfo("查询容量规划列表", "queryServerCapacityList", middleware.LIST, middleware.INFO)), server.CapacityList)
			// 容量计算
			serverGroup.GET("/capacity/count", middleware.OperatorLog(DefaultEventOpInfo("容量计算", "countServerCapacityList", middleware.LIST, middleware.INFO)), server.CapacityCount)
			// 保存容量规划
			serverGroup.POST("/capacity/save", middleware.OperatorLog(DefaultEventOpInfo("查询容量规划列表", "saveServerCapacityList", middleware.LIST, middleware.INFO)), server.SaveCapacity)
			// 下载服务器规划清单
			serverGroup.GET("/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载服务器规划清单", "downloadServerList", middleware.EXPORT, middleware.INFO)), server.Download)
			// 查询服务器上架表
			serverGroup.GET("/shelve/list", middleware.OperatorLog(DefaultEventOpInfo("查询服务器上架表", "getServerShelveList", middleware.LIST, middleware.INFO)), server.ListServerShelvePlanning)
			// 下载服务器上架表模板
			serverGroup.GET("/shelve/download/template/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载服务器上架表模板", "downloadServerShelveTemplate", middleware.EXPORT, middleware.INFO)), server.DownloadServerShelveTemplate)
			// 上传服务器上架表
			serverGroup.POST("/shelve/upload/:planId", middleware.OperatorLog(DefaultEventOpInfo("上传服务器上架表", "uploadServerShelve", middleware.IMPORT, middleware.INFO)), server.UploadShelve)
			// 保存服务器规划表
			serverGroup.POST("/shelve/planning/save", middleware.OperatorLog(DefaultEventOpInfo("保存服务器规划表", "saveServerPlanning", middleware.UPDATE, middleware.INFO)), server.SaveServerPlanning)
			// 保存服务器上架表
			serverGroup.POST("/shelve/save", middleware.OperatorLog(DefaultEventOpInfo("保存服务器上架表", "saveServerShelve", middleware.UPDATE, middleware.INFO)), server.SaveServerShelve)
			// 下载服务器上架清单
			serverGroup.GET("/shelve/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载服务器上架清单", "downloadServerShelve", middleware.EXPORT, middleware.INFO)), server.DownloadServerShelve)
		}

		// 网络规划
		networkGroup := api.Group("/network")
		{
			// 厂商列表查询
			networkGroup.GET("/brands", middleware.OperatorLog(DefaultEventOpInfo("厂商列表查询", "getBrandsByPlanId", middleware.LIST, middleware.INFO)), network_device.GetBrandsByPlanId)
			// 根据方案id获取网络设备规划信息
			networkGroup.GET("/plan/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id获取网络设备规划信息", "getDevicePlanByPlanId", middleware.GET, middleware.INFO)), network_device.GetDevicePlanByPlanId)
			// 获取网络设备清单
			networkGroup.POST("/device/list", middleware.OperatorLog(DefaultEventOpInfo("获取网络设备清单", "listNetworkDevices", middleware.LIST, middleware.INFO)), network_device.ListNetworkDevices)
			// 保存网络设备清单
			networkGroup.POST("/device/save", middleware.OperatorLog(DefaultEventOpInfo("保存网络设备清单", "saveDeviceList", middleware.CREATE, middleware.INFO)), network_device.SaveDeviceList)
			// 下载网络设备清单
			networkGroup.GET("/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载网络设备清单", "networkDeviceListDownload", middleware.EXPORT, middleware.INFO)), network_device.NetworkDeviceListDownload)
			// 查询网络设备上架列表
			networkGroup.GET("/shelve/list", middleware.OperatorLog(DefaultEventOpInfo("查询网络设备上架信息", "getNetworkShelveList", middleware.LIST, middleware.INFO)), network_device.ListNetworkShelve)
			// 下载网络设备上架模板
			networkGroup.GET("/shelve/download/template/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载网络设备上架模板", "downloadNetworkShelve", middleware.EXPORT, middleware.INFO)), network_device.DownloadNetworkShelveTemplate)
			// 上传网络设备上架表
			networkGroup.POST("/shelve/upload/:planId", middleware.OperatorLog(DefaultEventOpInfo("上传网络设备上架表", "uploadNetworkShelve", middleware.IMPORT, middleware.INFO)), network_device.UploadShelve)
			// 保存网络设备上架表
			networkGroup.POST("/shelve/save", middleware.OperatorLog(DefaultEventOpInfo("保存网络设备上架表", "saveNetworkShelve", middleware.UPDATE, middleware.INFO)), network_device.SaveShelve)
			// 下载网络设备上架清单
			networkGroup.GET("/shelve/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("保存网络设备上架清单", "saveNetworkShelve", middleware.EXPORT, middleware.INFO)), network_device.DownloadNetworkShelve)
		}

		// IP需求
		ipDemandGroup := api.Group("/ipDemand")
		{
			// 下载IP需求表
			ipDemandGroup.GET("/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载IP需求表", "ipDemandListDownload", middleware.EXPORT, middleware.INFO)), ip_demand.IpDemandListDownload)
			// 查询IP规划列表
			ipDemandGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询IP规划列表", "getIpDemandList", middleware.EXPORT, middleware.INFO)), ip_demand.List)
			// 上传IP需求表
			ipDemandGroup.POST("/upload/:planId", middleware.OperatorLog(DefaultEventOpInfo("上传IP需求表", "uploadIpDemand", middleware.IMPORT, middleware.INFO)), ip_demand.Upload)
			// 保存IP需求表
			ipDemandGroup.POST("/save", middleware.OperatorLog(DefaultEventOpInfo("保存IP需求表", "saveIpDemand", middleware.IMPORT, middleware.INFO)), ip_demand.Save)
		}

		// 云产品
		cloudProduct := api.Group("/cloud/product")
		{
			cloudProduct.GET("/version/list", middleware.OperatorLog(DefaultEventOpInfo("根据项目id查询云产品版本列表", "listCloudProductVersion", middleware.LIST, middleware.INFO)), cloud_product.ListVersion)
			cloudProduct.GET("/baseline/list", middleware.OperatorLog(DefaultEventOpInfo("根据版本id查询云产品基线列表", "listCloudProductBaseline", middleware.LIST, middleware.INFO)), cloud_product.ListCloudProductBaseline)
			cloudProduct.POST("/save", middleware.OperatorLog(DefaultEventOpInfo("保存用户选择的云产品", "saveCloudProduct", middleware.CREATE, middleware.INFO)), cloud_product.Save)
			cloudProduct.GET("/list/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id获取用户选择的云产品清单", "listCloudProduct", middleware.LIST, middleware.INFO)), cloud_product.List)
			cloudProduct.GET("/export/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载云服务规格清单", "exportCloudProduct", middleware.EXPORT, middleware.INFO)), cloud_product.Export)
		}

		// 机房规划
		machineRoomGroup := api.Group("/machineRoom")
		{
			machineRoomGroup.GET("/list/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id查询机房信息", "getMachineRoomByPlanId", middleware.GET, middleware.INFO)), machine_room.GetMachineRoomByPlanId)
			machineRoomGroup.PUT("/:planId", middleware.OperatorLog(DefaultEventOpInfo("修改机房信息", "updateMachineRoom", middleware.UPDATE, middleware.INFO)), machine_room.UpdateMachineRoom)
			machineRoomGroup.GET("/download", middleware.OperatorLog(DefaultEventOpInfo("下载机房勘察模版", "downloadCabinetTemplate", middleware.EXPORT, middleware.INFO)), machine_room.DownloadCabinetTemplate)
			machineRoomGroup.POST("/import", middleware.OperatorLog(DefaultEventOpInfo("导入机房勘察文件", "importCabinet", middleware.IMPORT, middleware.INFO)), machine_room.ImportCabinet)
			machineRoomGroup.GET("/cabinet/page", middleware.OperatorLog(DefaultEventOpInfo("机房规划机柜列表查询", "pageCabinet", middleware.LIST, middleware.INFO)), machine_room.PageCabinets)
		}

		// 全局配置
		globalConfigGroup := api.Group("/globalConfig")
		{
			globalConfigGroup.GET("/vlanId/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id查询vlan id配置信息", "getVlanIdConfigByPlanId", middleware.GET, middleware.INFO)), global_config.GetVlanIdConfigByPlanId)
			globalConfigGroup.POST("/vlanId", middleware.OperatorLog(DefaultEventOpInfo("新增vlan id配置信息", "createVlanIdConfig", middleware.CREATE, middleware.INFO)), global_config.CreateVlanIdConfig)
			// globalConfigGroup.PUT("/vlanId/:id", middleware.OperatorLog(DefaultEventOpInfo("修改vlan id配置信息", "updateVlanIdConfig", middleware.UPDATE, middleware.INFO)), global_config.UpdateVlanIdConfig)
			globalConfigGroup.GET("/cell/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id查询集群配置信息", "getCellConfigByPlanId", middleware.GET, middleware.INFO)), global_config.GetCellConfigByPlanId)
			globalConfigGroup.POST("/cell", middleware.OperatorLog(DefaultEventOpInfo("新增集群配置信息", "createCellConfig", middleware.CREATE, middleware.INFO)), global_config.CreateCellConfig)
			// globalConfigGroup.PUT("/cell/:id", middleware.OperatorLog(DefaultEventOpInfo("修改集群配置信息", "updateCellConfig", middleware.UPDATE, middleware.INFO)), global_config.UpdateCellConfig)
			globalConfigGroup.GET("/routePlanning/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id查询路由规划配置信息", "getRoutePlanningConfigByPlanId", middleware.GET, middleware.INFO)), global_config.GetRoutePlanningConfigByPlanId)
			globalConfigGroup.POST("/routePlanning", middleware.OperatorLog(DefaultEventOpInfo("新增路由规划配置信息", "createRoutePlanningConfig", middleware.CREATE, middleware.INFO)), global_config.CreateRoutePlanningConfig)
			// globalConfigGroup.PUT("/routePlanning/:id", middleware.OperatorLog(DefaultEventOpInfo("修改路由规划配置信息", "updateRoutePlanningConfig", middleware.UPDATE, middleware.INFO)), global_config.UpdateRoutePlanningConfig)
			globalConfigGroup.GET("/largeNetwork/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id查询大网网段配置信息", "getLargeNetworkConfigByPlanId", middleware.GET, middleware.INFO)), global_config.GetLargeNetworkConfigByPlanId)
			globalConfigGroup.POST("/largeNetwork", middleware.OperatorLog(DefaultEventOpInfo("新增大网网段配置信息", "createLargeNetworkPlanningConfig", middleware.CREATE, middleware.INFO)), global_config.CreateLargeNetworkConfig)
			// globalConfigGroup.PUT("/largeNetwork/:id", middleware.OperatorLog(DefaultEventOpInfo("修改大网网段配置信息", "updateLargeNetworkConfig", middleware.UPDATE, middleware.INFO)), global_config.UpdateLargeNetworkConfig)
			globalConfigGroup.POST("/complete/:planId", middleware.OperatorLog(DefaultEventOpInfo("全局配置完成规划", "completeGlobalConfig", middleware.CREATE, middleware.INFO)), global_config.CompleteGlobalConfig)
			globalConfigGroup.GET("/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载规划文件", "downloadPlanningFile", middleware.EXPORT, middleware.INFO)), global_config.DownloadPlanningFile)
		}
	}

	innerApi := engine.Group(innerPrefix)
	{
		// baseline
		baselineGroup := innerApi.Group("/baseline")
		{
			// 导入版本基线
			baselineGroup.POST("/import", middleware.OperatorLog(DefaultEventOpInfo("导入版本基线", "importBaseline", middleware.IMPORT, middleware.INFO)), baseline.Import)
		}
	}
	// 枚举配置表
	configGroup := api.Group("/config")
	{
		configGroup.GET("/:code", middleware.OperatorLog(DefaultEventOpInfo("查询枚举配置表", "queryConfig", middleware.LIST, middleware.INFO)), config_item.List)
	}
}

func DefaultEventOpInfo(actionDisplayName string, actionCode string, actionType middleware.ActionType, level middleware.EventLevel) middleware.LogConf {
	return middleware.LogConf{
		ServiceCode:        "planning-manage",
		ServiceDisplayName: "规划系统",
		UserSecretFile:     constant.UserSecretPrivateKeyPath,
		ActionCode:         actionCode,
		ActionDisplayName:  actionDisplayName,
		ActionType:         actionType,
		EventLevel:         level,
		RequestRegion:      os.Getenv(constant.EnvRegion),
		Extended:           "",
		RequestDescription: "",
	}
}

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		session := sessions.Default(context)
		currentUserIdInterface := session.Get("userId")
		if currentUserIdInterface == nil {
			log.Errorf("[Auth] invalid authorized")
			result.Failure(context, errorcodes.InvalidAuthorized, http.StatusUnauthorized)
			return
		}
		sessionAgeStr := os.Getenv("SESSION_AGE")
		sessionAge, _ := strconv.Atoi(sessionAgeStr)
		session.Options(sessions.Options{MaxAge: sessionAge, Path: "/"})
		session.Save()
		context.Set(constant.CurrentUserId, currentUserIdInterface.(string))
	}
}
