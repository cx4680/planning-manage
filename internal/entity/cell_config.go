package entity

import "time"

const CellConfigTable = "cell_config"

type CellConfig struct {
	Id                       int64     `gorm:"column:id" json:"id"`                                                // 主键Id
	PlanId                   int64     `gorm:"column:plan_id" json:"planId"`                                       // 方案Id
	BizRegionAbbr            string    `gorm:"column:biz_region_abbr" json:"bizRegionAbbr"`                        // 业务区域缩写
	CellSelfMgt              int       `gorm:"column:cell_self_mgt" json:"cellSelfMgt"`                            // 集群自纳管，0：否，1：是
	MgtGlobalDnsRootDomain   string    `gorm:"column:mgt_global_dns_root_domain" json:"mgtGlobalDnsRootDomain"`    // 管理网全局DNS根域
	GlobalDnsSvcAddress      string    `gorm:"column:global_dns_svc_address" json:"globalDnsSvcAddress"`           // 全局DNS服务地址
	CellVip                  string    `gorm:"column:cell_vip" json:"cellVip"`                                     // 集群Vip
	CellVipIpv6              string    `gorm:"column:cell_vip_ipv6" json:"cellVipIpv6"`                            // 集群Vip-IPV6地址
	ExternalNtpIp            string    `gorm:"column:external_ntp_ip" json:"externalNtpIp"`                        // 外部时钟源IP（多个时钟源以逗号分隔）
	NetworkMode              int       `gorm:"column:network_mode" json:"networkMode"`                             // 组网模式，0：标准模式，1：纯二层组网模式
	CellContainerNetwork     string    `gorm:"column:cell_container_network" json:"cellContainerNetwork"`          // 集群网络配置-集群容器网
	CellContainerNetworkIpv6 string    `gorm:"column:cell_container_network_ipv6" json:"cellContainerNetworkIpv6"` // 集群网络配置-集群容器网IPV6
	CellSvcNetwork           string    `gorm:"column:cell_svc_network" json:"cellSvcNetwork"`                      // 集群网络配置-集群服务网
	CellSvcNetworkIpv6       string    `gorm:"column:cell_svc_network_ipv6" json:"CellSvcNetworkIpv6"`             // 集群网络配置-集群服务网IPV6
	AddCellNodeSshPublicKey  string    `gorm:"column:add_cell_node_ssh_public_key" json:"addCellNodeSshPublicKey"` // 添加集群节点SSH访问公钥
	CreateUserId             string    `gorm:"column:create_user_id" json:"createUserId"`                          // 创建人id
	CreateTime               time.Time `gorm:"column:create_time" json:"createTime"`                               // 创建时间
	UpdateUserId             string    `gorm:"column:update_user_id" json:"updateUserId"`                          // 更新人id
	UpdateTime               time.Time `gorm:"column:update_time" json:"updateTime"`                               // 更新时间
}

func (entity *CellConfig) TableName() string {
	return CellConfigTable
}
