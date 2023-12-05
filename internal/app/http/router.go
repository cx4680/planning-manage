package http

import (
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/cloud_product"
	"code.cestc.cn/ccos/common/planning-manage/internal/svc/server"
	"os"

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

	"code.cestc.cn/ccos/common/planning-manage/internal/svc/baseline"
	"github.com/gin-gonic/gin"

	"code.cestc.cn/ccos/common/planning-manage/internal/api/constant"

	"code.cestc.cn/ccos/cnm/ops-base/opshttp/middleware"
)

const (
	apiPrefix   = "/api/planning-manage/v1"
	innerPrefix = "/api/planning-manage/v1/inner"
)

func Router(engine *gin.Engine) {

	v1 := engine.Group(apiPrefix)
	{
		// user
		userGroup := v1.Group("/user")
		{
			// 查询操作列表
			userGroup.POST("/login", middleware.OperatorLog(DefaultEventOpInfo("登录", "login", middleware.OPERATE, middleware.INFO)), user.Login)
			// 新增操作
			userGroup.GET("/logout", middleware.OperatorLog(DefaultEventOpInfo("登出", "logout", middleware.OPERATE, middleware.INFO)), user.Logout)
			// 根据名称查询ldap用户
			userGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("根据关键字查询用户", "listByName", middleware.OPERATE, middleware.INFO)), user.ListByName)
		}

		customerGroup := v1.Group("/customer")
		{
			// 分页查询客户列表
			customerGroup.POST("/page", middleware.OperatorLog(DefaultEventOpInfo("分页查询客户列表", "queryCustomerByPage", middleware.LIST, middleware.INFO)), customer.Page)
			// 获取客户列表
			//customerGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("获取客户列表", "listCustomer", middleware.LIST, middleware.INFO)), customer.List)
			// 根据id获取客户
			customerGroup.GET("/:id", middleware.OperatorLog(DefaultEventOpInfo("根据id获取客户", "queryCustomerById", middleware.GET, middleware.INFO)), customer.GetById)
			// 创建客户
			customerGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建客户", "createCustomer", middleware.CREATE, middleware.INFO)), customer.Create)
			// 根据id修改客户
			customerGroup.POST("/update", middleware.OperatorLog(DefaultEventOpInfo("根据id修改客户", "editCustomer", middleware.UPDATE, middleware.INFO)), customer.Update)
			// 根据id删除客户
			//customerGroup.DELETE("/delete/:id", middleware.OperatorLog(DefaultEventOpInfo("根据id删除客户", "deleteCustomer", middleware.DELETE, middleware.INFO)), customer.Delete)
		}

		// 云平台
		cloudPlatformGroup := v1.Group("/platform")
		{
			// 查询云平台列表
			cloudPlatformGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询云平台列表", "queryProjectList", middleware.LIST, middleware.INFO)), cloud_platform.List)
			// 创建云平台
			cloudPlatformGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建云平台", "createProject", middleware.CREATE, middleware.INFO)), cloud_platform.Create)
			// 根据id修改云平台
			cloudPlatformGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改云平台", "deleteProject", middleware.UPDATE, middleware.INFO)), cloud_platform.Update)
		}

		// region
		regionGroup := v1.Group("/region")
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
		azGroup := v1.Group("/az")
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
		cellGroup := v1.Group("/cell")
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

		// project
		projectGroup := v1.Group("/project")
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

		// plan
		planGroup := v1.Group("/plan")
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

		//服务器规划
		serverGroup := v1.Group("/server")
		{
			// 查询服务器规划列表
			serverGroup.GET("/list", middleware.OperatorLog(DefaultEventOpInfo("查询服务器列表", "queryServerList", middleware.LIST, middleware.INFO)), server.List)
			// 创建服务器规划
			serverGroup.POST("/create", middleware.OperatorLog(DefaultEventOpInfo("创建服务器", "createServerList", middleware.LIST, middleware.INFO)), server.Create)
			// 修改服务器规划
			serverGroup.PUT("/update/:id", middleware.OperatorLog(DefaultEventOpInfo("修改服务器", "updateServerList", middleware.LIST, middleware.INFO)), server.Update)
			// 查询服务器架构列表
			serverGroup.GET("/arch/list", middleware.OperatorLog(DefaultEventOpInfo("查询服务器架构列表", "queryServerArchList", middleware.LIST, middleware.INFO)), server.ArchList)
		}

		// baseline
		baselineGroup := v1.Group("/baseline")
		{
			// 导入版本基线
			baselineGroup.POST("/import", middleware.OperatorLog(DefaultEventOpInfo("导入版本基线", "importBaseline", middleware.IMPORT, middleware.INFO)), baseline.Import)
		}

		// network
		networkGroup := v1.Group("/network")
		{
			// 计算机柜数量
			//networkGroup.GET("/box/count", middleware.OperatorLog(DefaultEventOpInfo("计算机柜数量", "getCountBoxNum", middleware.GET, middleware.INFO)), network_device.GetCountBoxNum)
			// 厂商列表查询
			networkGroup.GET("/brands/:planId", middleware.OperatorLog(DefaultEventOpInfo("厂商列表查询", "getBrandsByPlanId", middleware.LIST, middleware.INFO)), network_device.GetBrandsByPlanId)
			// 根据方案id获取网络设备规划信息
			networkGroup.GET("/plan/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id获取网络设备规划信息", "getDevicePlanByPlanId", middleware.GET, middleware.INFO)), network_device.GetDevicePlanByPlanId)
			// 获取网络设备清单
			networkGroup.POST("/device/list", middleware.OperatorLog(DefaultEventOpInfo("获取网络设备清单", "listNetworkDevices", middleware.LIST, middleware.INFO)), network_device.ListNetworkDevices)
			// 获取网络设备清单
			networkGroup.POST("/device/save", middleware.OperatorLog(DefaultEventOpInfo("保存网络设备清单", "saveDeviceList", middleware.CREATE, middleware.INFO)), network_device.SaveDeviceList)
		}

		// ipDemand
		ipDemandGroup := v1.Group("/ipDemand")
		{
			// 下载IP需求表
			ipDemandGroup.GET("/download/:planId", middleware.OperatorLog(DefaultEventOpInfo("下载IP需求表", "ipDemandListDownload", middleware.EXPORT, middleware.INFO)), ip_demand.IpDemandListDownload)
		}

		// cloudProduct
		cloudProduct := v1.Group("/cloud/product")
		{
			cloudProduct.GET("/version/list", middleware.OperatorLog(DefaultEventOpInfo("根据项目id查询云产品版本列表", "listCloudProductVersion", middleware.LIST, middleware.INFO)), cloud_product.ListVersion)
			cloudProduct.GET("/baseline/list", middleware.OperatorLog(DefaultEventOpInfo("根据版本id查询云产品基线列表", "listCloudProductBaseline", middleware.LIST, middleware.INFO)), cloud_product.ListCloudProductBaseline)
			cloudProduct.POST("/save", middleware.OperatorLog(DefaultEventOpInfo("保存用户选择的云产品", "saveCloudProduct", middleware.CREATE, middleware.INFO)), cloud_product.Save)
			cloudProduct.GET("/list/:planId", middleware.OperatorLog(DefaultEventOpInfo("根据方案id获取用户选择的云产品清单", "listCloudProduct", middleware.LIST, middleware.INFO)), cloud_product.List)
		}
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
