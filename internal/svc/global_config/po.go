package global_config

import (
	"time"

	"code.cestc.cn/ccos/common/planning-manage/internal/entity"
)

type VlanIdConfigRequest struct {
	PlanId             int64  `form:"planId" json:"planId"`                         // 方案Id
	InBandMgtVlanId    string `form:"inBandMgtVlanId" json:"inBandMgtVlanId"`       // 带内管理Vlan ID
	LocalStorageVlanId string `form:"localStorageVlanId" json:"localStorageVlanId"` // 本地存储网Vlan ID
	BizIntranetVlanId  string `form:"bizIntranetVlanId" json:"bizIntranetVlanId"`   // 业务内网Vlan ID
}

type CellConfigReq struct {
	PlanId                   int64  `form:"planId" json:"planId"`                                     // 方案Id
	RegionCode               string `form:"regionCode" json:"regionCode"`                             // 区域编码
	RegionType               string `form:"regionType" json:"regionType"`                             // 区域类型
	BizRegionAbbr            string `form:"bizRegionAbbr" json:"bizRegionAbbr"`                       // 业务区域缩写
	AzCode                   string `form:"azCode" json:"azCode"`                                     // 可用区编码
	CellType                 string `form:"cellType" json:"cellType"`                                 // 集群类型
	CellName                 string `form:"cellName" json:"cellName"`                                 // 集群名称
	CellSelfMgt              int    `form:"cellSelfMgt" json:"cellSelfMgt"`                           // 集群自纳管，0：否，1：是
	MgtGlobalDnsRootDomain   string `form:"mgtGlobalDnsRootDomain" json:"mgtGlobalDnsRootDomain"`     // 管理网全局DNS根域
	GlobalDnsSvcAddress      string `form:"globalDnsSvcAddress" json:"globalDnsSvcAddress"`           // 全局DNS服务地址
	CellVip                  string `form:"cellVip" json:"cellVip"`                                   // 集群Vip
	CellVipIpv6              string `form:"cellVipIpv6" json:"cellVipIpv6"`                           // 集群Vip-IPV6地址
	ExternalNtpIp            string `form:"externalNtpIp" json:"externalNtpIp"`                       // 外部时钟源IP（多个时钟源以逗号分隔）
	NetworkMode              int    `form:"networkMode" json:"networkMode"`                           // 组网模式，0：标准模式，1：纯二层组网模式
	CellContainerNetwork     string `form:"cellContainerNetwork" json:"cellContainerNetwork"`         // 集群网络配置-集群容器网
	CellContainerNetworkIpv6 string `form:"cellContainerNetworkIpv6" json:"cellContainerNetworkIpv6"` // 集群网络配置-集群容器网IPV6
	CellSvcNetwork           string `form:"cellSvcNetwork" json:"cellSvcNetwork"`                     // 集群网络配置-集群服务网
	CellSvcNetworkIpv6       string `form:"CellSvcNetworkIpv6" json:"CellSvcNetworkIpv6"`             // 集群网络配置-集群服务网IPV6
	AddCellNodeSshPublicKey  string `form:"addCellNodeSshPublicKey" json:"addCellNodeSshPublicKey"`   // 添加集群节点SSH访问公钥
}

type RegionAzCell struct {
	RegionId   int64  `form:"regionId" json:"regionId"`     // 区域id
	RegionCode string `form:"regionCode" json:"regionCode"` // 区域编码
	RegionType string `form:"regionType" json:"regionType"` // 区域类型
	AzId       int64  `form:"azId" json:"azId"`             // 可用区id
	AzCode     string `form:"azCode" json:"azCode"`         // 可用区编码
	CellId     int64  `form:"cellId" json:"cellId"`         // 集群id
	CellType   string `form:"cellType" json:"cellType"`     // 集群类型
	CellName   string `form:"cellName" json:"cellName"`     // 集群名称
}

type CellConfigResp struct {
	Id                       int64  `form:"id" json:"id"`                                             // 主键id
	PlanId                   int64  `form:"planId" json:"planId"`                                     // 方案Id
	RegionCode               string `form:"regionCode" json:"regionCode"`                             // 区域编码
	RegionType               string `form:"regionType" json:"regionType"`                             // 区域类型
	BizRegionAbbr            string `form:"bizRegionAbbr" json:"bizRegionAbbr"`                       // 业务区域缩写
	AzCode                   string `form:"azCode" json:"azCode"`                                     // 可用区编码
	CellType                 string `form:"cellType" json:"cellType"`                                 // 集群类型
	CellName                 string `form:"cellName" json:"cellName"`                                 // 集群名称
	CellSelfMgt              int    `form:"cellSelfMgt" json:"cellSelfMgt"`                           // 集群自纳管，0：否，1：是
	MgtGlobalDnsRootDomain   string `form:"mgtGlobalDnsRootDomain" json:"mgtGlobalDnsRootDomain"`     // 管理网全局DNS根域
	GlobalDnsSvcAddress      string `form:"globalDnsSvcAddress" json:"globalDnsSvcAddress"`           // 全局DNS服务地址
	CellVip                  string `form:"cellVip" json:"cellVip"`                                   // 集群Vip
	CellVipIpv6              string `form:"cellVipIpv6" json:"cellVipIpv6"`                           // 集群Vip-IPV6地址
	ExternalNtpIp            string `form:"externalNtpIp" json:"externalNtpIp"`                       // 外部时钟源IP（多个时钟源以逗号分隔）
	NetworkMode              int    `form:"networkMode" json:"networkMode"`                           // 组网模式，0：标准模式，1：纯二层组网模式
	CellContainerNetwork     string `form:"cellContainerNetwork" json:"cellContainerNetwork"`         // 集群网络配置-集群容器网
	CellContainerNetworkIpv6 string `form:"cellContainerNetworkIpv6" json:"cellContainerNetworkIpv6"` // 集群网络配置-集群容器网IPV6
	CellSvcNetwork           string `form:"cellSvcNetwork" json:"cellSvcNetwork"`                     // 集群网络配置-集群服务网
	CellSvcNetworkIpv6       string `form:"CellSvcNetworkIpv6" json:"CellSvcNetworkIpv6"`             // 集群网络配置-集群服务网IPV6
	AddCellNodeSshPublicKey  string `form:"addCellNodeSshPublicKey" json:"addCellNodeSshPublicKey"`   // 添加集群节点SSH访问公钥
}

type RoutePlanningConfigReq struct {
	PlanId                      int64  `form:"planId" json:"planId"`                                           // 方案Id
	DeployUseBgp                int    `form:"deployUseBgp" json:"deployUseBgp"`                               // 使用BGP部署，0：否，1：是
	DeployMachSwitchSelfNum     string `form:"deployMachSwitchSelfNum" json:"deployMachSwitchSelfNum"`         // 部署机所在交换机自治号
	DeployMachSwitchIp          string `form:"deployMachSwitchIp" json:"deployMachSwitchIp"`                   // 部署机所在交换机IP（多个IP以逗号分隔）
	SvcExternalAccessAddress    string `form:"svcExternalAccessAddress" json:"svcExternalAccessAddress"`       // 服务外部访问地址
	BgpNeighbor                 string `form:"bgpNeighbor" json:"bgpNeighbor"`                                 // BGP邻居
	CellDnsSvcAddress           string `form:"cellDnsSvcAddress" json:"cellDnsSvcAddress"`                     // 集群DNS服务地址
	RegionDnsSvcAddress         string `form:"regionDnsSvcAddress" json:"regionDnsSvcAddress"`                 // Region DNS服务地址
	OpsCenterIp                 string `form:"opsCenterIp" json:"opsCenterIp"`                                 // 运维中心访问IP
	OpsCenterIpv6               string `form:"opsCenterIpv6" json:"opsCenterIpv6"`                             // 运维中心访问IPV6地址
	OpsCenterPort               string `form:"opsCenterPort" json:"opsCenterPort"`                             // 运维中心访问端口
	OpsCenterDomain             string `form:"opsCenterDomain" json:"opsCenterDomain"`                         // 运维中心访问域名
	OperationCenterIp           string `form:"operationCenterIp" json:"operationCenterIp"`                     // 运营中心访问IP
	OperationCenterIpv6         string `form:"operationCenterIpv6" json:"operationCenterIpv6"`                 // 运营中心访问IPV6地址
	OperationCenterPort         string `form:"operationCenterPort" json:"operationCenterPort"`                 // 运营中心访问端口
	OperationCenterDomain       string `form:"operationCenterDomain" json:"operationCenterDomain"`             // 运营中心访问域名
	OpsCenterInitUserName       string `form:"opsCenterInitUserName" json:"opsCenterInitUserName"`             // 运维中心初始化用户配置-用户名
	OpsCenterInitUserPwd        string `form:"opsCenterInitUserPwd" json:"opsCenterInitUserPwd"`               // 运维中心初始化用户配置-密码
	OperationCenterInitUserName string `form:"operationCenterInitUserName" json:"operationCenterInitUserName"` // 运营中心初始化用户配置-用户名
	OperationCenterInitUserPwd  string `form:"operationCenterInitUserPwd" json:"operationCenterInitUserPwd"`   // 运营中心初始化用户配置-密码
}

func ConvertRoutePlanningReq2Entity(userId string, now time.Time, request RoutePlanningConfigReq) entity.RoutePlanningConfig {
	routePlanningConfig := entity.RoutePlanningConfig{
		PlanId:                      request.PlanId,
		DeployUseBgp:                request.DeployUseBgp,
		DeployMachSwitchSelfNum:     request.DeployMachSwitchSelfNum,
		DeployMachSwitchIp:          request.DeployMachSwitchIp,
		SvcExternalAccessAddress:    request.SvcExternalAccessAddress,
		BgpNeighbor:                 request.BgpNeighbor,
		CellDnsSvcAddress:           request.CellDnsSvcAddress,
		RegionDnsSvcAddress:         request.RegionDnsSvcAddress,
		OpsCenterIp:                 request.OpsCenterIp,
		OpsCenterIpv6:               request.OpsCenterIpv6,
		OpsCenterPort:               request.OpsCenterPort,
		OpsCenterDomain:             request.OpsCenterDomain,
		OperationCenterIp:           request.OperationCenterIp,
		OperationCenterIpv6:         request.OperationCenterIpv6,
		OperationCenterPort:         request.OperationCenterPort,
		OperationCenterDomain:       request.OperationCenterDomain,
		OpsCenterInitUserName:       request.OpsCenterInitUserName,
		OpsCenterInitUserPwd:        request.OpsCenterInitUserPwd,
		OperationCenterInitUserName: request.OperationCenterInitUserName,
		OperationCenterInitUserPwd:  request.OperationCenterInitUserPwd,
		UpdateUserId:                userId,
		UpdateTime:                  now,
	}
	return routePlanningConfig
}

type LargeNetworkSegmentConfigReq struct {
	PlanId                         int64  `form:"planId" json:"planId"`                                                 // 方案Id
	StorageNetworkSegmentRoute     string `form:"storageNetworkSegmentRoute" json:"storageNetworkSegmentRoute"`         // 存储前端网规划网段明细路由
	BizIntranetNetworkSegmentRoute string `form:"bizIntranetNetworkSegmentRoute" json:"bizIntranetNetworkSegmentRoute"` // 业务内网规划网段明细路由
	BizExternalLargeNetworkSegment string `form:"bizExternalLargeNetworkSegment" json:"bizExternalLargeNetworkSegment"` // 业务外网大网网段
	BmcNetworkSegmentRoute         string `form:"bmcNetworkSegmentRoute" json:"bmcNetworkSegmentRoute"`                 // bmc规划网段明细路由
}

type GlobalConfigExcel struct {
	InBandMgtVlanId                string `excel:"cellPosition:B2;" json:"inBandMgtVlanId"`                // 带内管理Vlan ID
	LocalStorageVlanId             string `excel:"cellPosition:B3;" json:"localStorageVlanId"`             // 本地存储网Vlan ID
	BizIntranetVlanId              string `excel:"cellPosition:B4;" json:"bizIntranetVlanId"`              // 业务内网Vlan ID
	RegionCode                     string `excel:"cellPosition:D2;" json:"regionCode"`                     // 区域编码
	AzCode                         string `excel:"cellPosition:D3;" json:"azCode"`                         // 可用区编码
	RegionType                     string `excel:"cellPosition:D4;" json:"regionType"`                     // 区域类型
	CellType                       string `excel:"cellPosition:D5;" json:"cellType"`                       // 集群类型
	CellSelfMgt                    string `excel:"cellPosition:D6;" json:"cellSelfMgt"`                    // 集群自纳管，0：否，1：是
	CellName                       string `excel:"cellPosition:D7;" json:"cellName"`                       // 集群名称
	MgtGlobalDnsRootDomain         string `excel:"cellPosition:D8;" json:"mgtGlobalDnsRootDomain"`         // 管理网全局DNS根域
	GlobalDnsSvcAddress            string `excel:"cellPosition:D9;" json:"globalDnsSvcAddress"`            // 全局DNS服务地址
	DualStackDeploy                string `excel:"cellPosition:D10;" json:"dualStackDeploy"`               // 是否双栈交付
	CellVip                        string `excel:"cellPosition:D11;" json:"cellVip"`                       // 集群Vip
	CellVipIpv6                    string `excel:"cellPosition:D12;" json:"cellVipIpv6"`                   // 集群Vip-IPV6地址
	ExternalNtpIp                  string `excel:"cellPosition:D13;" json:"externalNtpIp"`                 // 外部时钟源IP（多个时钟源以逗号分隔）
	NetworkMode                    string `excel:"cellPosition:D14;" json:"networkMode"`                   // 组网模式，0：标准模式，1：纯二层组网模式
	CellContainerNetwork           string `excel:"cellPosition:D15;" json:"cellContainerNetwork"`          // 集群网络配置-集群容器网
	CellContainerNetworkIpv6       string `excel:"cellPosition:D16;" json:"cellContainerNetworkIpv6"`      // 集群网络配置-集群容器网IPV6
	CellSvcNetwork                 string `excel:"cellPosition:D17;" json:"cellSvcNetwork"`                // 集群网络配置-集群服务网
	CellSvcNetworkIpv6             string `excel:"cellPosition:D18;" json:"CellSvcNetworkIpv6"`            // 集群网络配置-集群服务网IPV6
	AddCellNodeSshPublicKey        string `excel:"cellPosition:D19;" json:"addCellNodeSshPublicKey"`       // 添加集群节点SSH访问公钥
	DeployUseBgp                   string `excel:"cellPosition:F2;" json:"deployUseBgp"`                   // 使用BGP部署，0：否，1：是
	DeployMachSwitchSelfNum        string `excel:"cellPosition:F3;" json:"deployMachSwitchSelfNum"`        // 部署机所在交换机自治号
	DeployMachSwitchIp             string `excel:"cellPosition:F4;" json:"deployMachSwitchIp"`             // 部署机所在交换机IP（多个IP以逗号分隔）
	SvcExternalAccessAddress       string `excel:"cellPosition:F5;" json:"svcExternalAccessAddress"`       // 服务外部访问地址
	BgpNeighbor                    string `excel:"cellPosition:F6;" json:"bgpNeighbor"`                    // BGP邻居
	CellDnsSvcAddress              string `excel:"cellPosition:F7;" json:"cellDnsSvcAddress"`              // 集群DNS服务地址
	RegionDnsSvcAddress            string `excel:"cellPosition:F8;" json:"regionDnsSvcAddress"`            // Region DNS服务地址
	OpsCenterIp                    string `excel:"cellPosition:F9;" json:"opsCenterIp"`                    // 运维中心访问IP
	OpsCenterIpv6                  string `excel:"cellPosition:F10;" json:"opsCenterIpv6"`                 // 运维中心访问IPV6地址
	OpsCenterPort                  string `excel:"cellPosition:F11;" json:"opsCenterPort"`                 // 运维中心访问端口
	OpsCenterDomain                string `excel:"cellPosition:F12;" json:"opsCenterDomain"`               // 运维中心访问域名
	OperationCenterIp              string `excel:"cellPosition:F13;" json:"operationCenterIp"`             // 运营中心访问IP
	OperationCenterIpv6            string `excel:"cellPosition:F14;" json:"operationCenterIpv6"`           // 运营中心访问IPV6地址
	OperationCenterPort            string `excel:"cellPosition:F15;" json:"operationCenterPort"`           // 运营中心访问端口
	OperationCenterDomain          string `excel:"cellPosition:F16;" json:"operationCenterDomain"`         // 运营中心访问域名
	OpsCenterInitUserName          string `excel:"cellPosition:F17;" json:"opsCenterInitUserName"`         // 运维中心初始化用户配置-用户名
	OpsCenterInitUserPwd           string `excel:"cellPosition:F18;" json:"opsCenterInitUserPwd"`          // 运维中心初始化用户配置-密码
	OperationCenterInitUserName    string `excel:"cellPosition:F19;" json:"operationCenterInitUserName"`   // 运营中心初始化用户配置-用户名
	OperationCenterInitUserPwd     string `excel:"cellPosition:F20;" json:"operationCenterInitUserPwd"`    // 运营中心初始化用户配置-密码
	StorageNetworkSegmentRoute     string `excel:"cellPosition:H2;" json:"storageNetworkSegmentRoute"`     // 存储前端网规划网段明细路由
	BizIntranetNetworkSegmentRoute string `excel:"cellPosition:H3;" json:"bizIntranetNetworkSegmentRoute"` // 业务内网规划网段明细路由
	BizExternalLargeNetworkSegment string `excel:"cellPosition:H4;" json:"bizExternalLargeNetworkSegment"` // 业务外网大网网段
	BmcNetworkSegmentRoute         string `excel:"cellPosition:H5;" json:"bmcNetworkSegmentRoute"`         // bmc规划网段明细路由
}

type GlobalConfigNetworkDeviceExcel struct {
	LogicalGrouping               string `excel:"name:交换机组名;index:0;" json:"LogicalGrouping"`                     // 网络设备逻辑分组
	MachineRoomAbbr               string `excel:"name:机房;index:1;" json:"machineRoomAbbr"`                        // 机房缩写
	CabinetNum                    string `excel:"name:机架;index:2;" json:"cabinetNum"`                             // 机柜编号
	SlotNum                       string `excel:"name:机位;index:3;" json:"slotNum"`                                // 机位槽位号
	CabinetAsw                    string `excel:"name:所管理的机架;index:4;" json:"cabinetAsw"`                         // 所管理的机架
	PxeSubnet                     string `excel:"name:PXE子网;index:5;" json:"pxeSubnet"`                           // PXE子网
	PxeSubnetRange                string `excel:"name:PXE子网范围;index:6;" json:"pxeSubnetRange"`                    // PXE子网范围
	PxeNetworkGateway             string `excel:"name:PXE网网关;index:7;" json:"pxeNetworkGateway"`                  // PXE网网关
	ManageSubnet                  string `excel:"name:管理网子网;index:8;" json:"manageSubnet"`                        // 管理网子网
	ManageNetworkGateway          string `excel:"name:管理网网关;index:9;" json:"manageNetworkGateway"`                // 管理网网关
	ManageIpv6Subnet              string `excel:"name:管理网IPV6子网;index:10;" json:"manageIpv6Subnet"`               // 管理网IPV6子网
	ManageIpv6NetworkGateway      string `excel:"name:管理网IPV6网关;index:11;" json:"manageIpv6NetworkGateway"`       // 管理网IPV6网关
	BizSubnet                     string `excel:"name:业务网子网;index:12;" json:"bizSubnet"`                          // 业务网子网
	BizNetworkGateway             string `excel:"name:业务网网关;index:13;" json:"bizNetworkGateway"`                  // 业务网网关
	StorageFrontNetwork           string `excel:"name:存储前端网;index:14;" json:"storageFrontNetwork"`                // 存储前端网
	StorageFrontNetworkGateway    string `excel:"name:存储前端网网关;index:15;" json:"storageFrontNetworkGateway"`       // 存储前端网网关
	BizExternalSubnet             string `excel:"name:业务外网子网;index:16;" json:"bizExternalSubnet"`                 // 业务外网子网
	BizExternalNetworkGateway     string `excel:"name:业务外网网关;index:17;" json:"bizExternalNetworkGateway"`         // 业务外网网关
	BmcSubnet                     string `excel:"name:bmc子网;index:18;" json:"bmcSubnet"`                          // bmc子网
	BmcNetworkGateway             string `excel:"name:bmc网关;index:19;" json:"bmcNetworkGateway"`                  // BMC网网关
	BizExternalIpv6Subnet         string `excel:"name:业务外网ipv6子网;index:20;" json:"bizExternalIpv6Subnet"`         // 业务外网ipv6子网
	BizExternalIpv6NetworkGateway string `excel:"name:业务外网ipv6网关;index:21;" json:"bizExternalIpv6NetworkGateway"` // 业务外网ipv6网关
}

type GlobalConfigServerExcel struct {
	Sn                       string `excel:"name:SN号;index:0;" json:"sn"`                               // SN号
	MachineRoomAbbr          string `excel:"name:机房;index:1;" json:"machineRoomAbbr"`                   // 机房缩写
	CabinetNum               string `excel:"name:机架;index:2;" json:"cabinetNum"`                        // 机柜编号
	SlotNum                  string `excel:"name:机位;index:3;" json:"slotNum"`                           // 机位槽位号
	KernelArch               string `excel:"name:内核架构;index:4;" json:"kernelArch"`                      // 内核架构
	HostName                 string `excel:"name:主机名;index:5;" json:"hostName"`                         // 主机名
	NetworkMode              string `excel:"name:网络模式;index:6;" json:"networkMode"`                     // 网络模式
	Role                     string `excel:"name:角色;index:7;" json:"role"`                              // 角色
	NodeTwoLayerNetwork      string `excel:"name:节点纯二层网络;index:8;" json:"nodeTwoLayerNetwork"`          // 节点纯二层网络
	HostGroup                string `excel:"name:主机分组;index:9;" json:"hostGroup"`                       // 主机分组
	OpenDpdk                 string `excel:"name:是否开启dpdk;index:10;" json:"openDpdk"`                   // 是否开启dpdk
	ManegeNetworkIP          string `excel:"name:管理网IP;index:11;" json:"manegeNetworkIP"`               // 管理网IP
	ManageNetworkIpv6IP      string `excel:"name:管理网IPV6;index:12;" json:"manageNetworkIpv6IP"`         // 管理网IPV6
	BizIntranetIP            string `excel:"name:业务内网IP;index:13;" json:"bizIntranetIP"`                // 业务内网IP
	BizExternalNetworkIP     string `excel:"name:业务外网IP;index:14;" json:"bizExternalNetworkIP"`         // 业务外网IP
	StorageNetworkIP         string `excel:"name:存储网IP;index:15;" json:"storageNetworkIP"`              // 存储网IP
	BmcNetworkIP             string `excel:"name:BMC网IP;index:16;" json:"bmcNetworkIP"`                 // BMC网IP
	ManageNetworkInterface   string `excel:"name:管理网网口;index:17;" json:"manageNetworkInterface"`        // 管理网网口
	ManageNetworkBondMode    string `excel:"name:管理网BOND模式;index:18;" json:"manageNetworkBondMode"`     // 管理网BOND模式
	BizIntranetInterface     string `excel:"name:业务内网网口;index:19;" json:"bizIntranetInterface"`         // 业务内网网口
	BizIntranetBondMode      string `excel:"name:业务内网BOND模式;index:20;" json:"bizIntranetBondMode"`      // 业务内网BOND模式
	StorageNetworkInterface  string `excel:"name:存储网网口;index:21;" json:"storageNetworkInterface"`       // 存储网网口
	StorageNetworkBondMode   string `excel:"name:存储网BOND模式;index:22;" json:"storageNetworkBondMode"`    // 存储网BOND模式
	BmcNetworkInterface      string `excel:"name:BMC网网口;index:23;" json:"bmcNetworkInterface"`          // BMC网网口
	BizExternalInterface     string `excel:"name:业务外网网口;index:24;" json:"bizExternalInterface"`         // 业务外网网口
	BizExternalBondMode      string `excel:"name:业务外网BOND模式;index:25;" json:"bizExternalBondMode"`      // 业务外网BOND模式
	BcmNetworkInterface      string `excel:"name:BMC网网口;index:26;" json:"bcmNetworkInterface"`          // BMC网网口
	BmcNetworkBondMode       string `excel:"name:BMC网BOND模式;index:27;" json:"bcmNetworkBondMode"`       // BMC网BOND模式
	PxeMacAddress            string `excel:"name:PXE-MAC地址;index:28;" json:"pxeMacAddress"`             // PXE-MAC地址
	BmcIp                    string `excel:"name:BMC-IP;index:29;" json:"bmcIp"`                        // BMC-IP
	BmcUserName              string `excel:"name:BMC用户名;index:30;" json:"bmcUserName"`                  // BMC用户名
	BmcPassword              string `excel:"name:BMC密码;index:31;" json:"bmcPassword"`                   // BMC密码
	BmcSecuritySuite         string `excel:"name:BMC安全套件;index:32;" json:"bmcSecuritySuite"`            // BMC安全套件
	SystemDiskRaidController string `excel:"name:系统盘RAID控制器;index:33;" json:"systemDiskRaidController"` // 系统盘RAID控制器
	SystemDiskRaidLevel      string `excel:"name:RAID等级;index:34;" json:"systemDiskRaidLevel"`          // 系统盘RAID等级
	SystemDisk               string `excel:"name:系统盘;index:35;" json:"systemDisk"`                      // 系统盘
	DataDiskRaidController   string `excel:"name:数据盘RAID控制器;index:36;" json:"dataDiskRaidController"`   // 数据盘RAID控制器
	DataDiskRaidLevel        string `excel:"name:RAID等级;index:37;" json:"dataDiskRaidLevel"`            // 数据盘RAID等级
	DataDisk                 string `excel:"name:数据盘;index:38;" json:"dataDisk"`                        // 数据盘
}
