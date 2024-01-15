package entity

import "time"

const RoutePlanningConfigTable = "route_planning_config"

type RoutePlanningConfig struct {
	Id                          int64     `gorm:"column:id" json:"id"`                                                       // 主键Id
	PlanId                      int64     `gorm:"column:plan_id" json:"planId"`                                              // 方案Id
	DeployUseBgp                int       `gorm:"column:deploy_use_bgp" json:"deployUseBgp"`                                 // 使用BGP部署，0：否，1：是
	DeployMachSwitchSelfNum     string    `gorm:"column:deploy_mach_switch_self_num" json:"deployMachSwitchSelfNum"`         // 部署机所在交换机自治号
	DeployMachSwitchIp          string    `gorm:"column:deploy_mach_switch_ip" json:"deployMachSwitchIp"`                    // 部署机所在交换机IP（多个IP以逗号分隔）
	SvcExternalAccessAddress    string    `gorm:"column:svc_external_access_address" json:"svcExternalAccessAddress"`        // 服务外部访问地址
	BgpNeighbor                 string    `gorm:"column:bgp_neighbor" json:"bgpNeighbor"`                                    // BGP邻居
	CellDnsSvcAddress           string    `gorm:"column:cell_dns_svc_address" json:"cellDnsSvcAddress"`                      // 集群DNS服务地址
	RegionDnsSvcAddress         string    `gorm:"column:region_dns_svc_address" json:"regionDnsSvcAddress"`                  // Region DNS服务地址
	OpsCenterIp                 string    `gorm:"column:ops_center_ip" json:"opsCenterIp"`                                   // 运维中心访问IP
	OpsCenterIpv6               string    `gorm:"column:ops_center_ipv6" json:"opsCenterIpv6"`                               // 运维中心访问IPV6地址
	OpsCenterPort               string    `gorm:"column:ops_center_port" json:"opsCenterPort"`                               // 运维中心访问端口
	OpsCenterDomain             string    `gorm:"column:ops_center_domain" json:"opsCenterDomain"`                           // 运维中心访问域名
	OperationCenterIp           string    `gorm:"column:operation_center_ip" json:"operationCenterIp"`                       // 运营中心访问IP
	OperationCenterIpv6         string    `gorm:"column:operation_center_ipv6" json:"operationCenterIpv6"`                   // 运营中心访问IPV6地址
	OperationCenterPort         string    `gorm:"column:operation_center_port" json:"operationCenterPort"`                   // 运营中心访问端口
	OperationCenterDomain       string    `gorm:"column:operation_center_domain" json:"operationCenterDomain"`               // 运营中心访问域名
	OpsCenterInitUserName       string    `gorm:"column:ops_center_init_user_name" json:"opsCenterInitUserName"`             // 运维中心初始化用户配置-用户名
	OpsCenterInitUserPwd        string    `gorm:"column:ops_center_init_user_pwd" json:"opsCenterInitUserPwd"`               // 运维中心初始化用户配置-密码
	OperationCenterInitUserName string    `gorm:"column:operation_center_init_user_name" json:"operationCenterInitUserName"` // 运营中心初始化用户配置-用户名
	OperationCenterInitUserPwd  string    `gorm:"column:operation_center_init_user_pwd" json:"operationCenterInitUserPwd"`   // 运营中心初始化用户配置-密码
	CreateUserId                string    `gorm:"column:create_user_id" json:"createUserId"`                                 // 创建人id
	CreateTime                  time.Time `gorm:"column:create_time" json:"createTime"`                                      // 创建时间
	UpdateUserId                string    `gorm:"column:update_user_id" json:"updateUserId"`                                 // 更新人id
	UpdateTime                  time.Time `gorm:"column:update_time" json:"updateTime"`                                      // 更新时间
}

func (entity *RoutePlanningConfig) TableName() string {
	return RoutePlanningConfigTable
}
