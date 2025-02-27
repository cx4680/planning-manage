CREATE TABLE `machine_room`
(
    `az_id`    bigint(20) NULL DEFAULT NULL COMMENT 'azId',
    `name`     varchar(255) NULL DEFAULT NULL COMMENT '机房名称',
    `abbr`     varchar(255) NULL DEFAULT NULL COMMENT '机房缩写',
    `province` varchar(50) NULL DEFAULT NULL COMMENT '省',
    `city`     varchar(50) NULL DEFAULT NULL COMMENT '市',
    `address`  varchar(50) NULL DEFAULT NULL COMMENT '地址',
    `sort`     int NULL DEFAULT NULL COMMENT '排序'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机房表';


CREATE TABLE `network_device_shelve`
(
    `id`                     bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `plan_id`                bigint(20) DEFAULT NULL COMMENT ' 方案id',
    `device_logical_id`      varchar(255) DEFAULT NULL COMMENT '网络设备逻辑ID',
    `device_id`              varchar(255) DEFAULT NULL COMMENT '网络设备ID',
    `sn`                     bigint(20) DEFAULT NULL COMMENT 'SN',
    `network_device_role_id` bigint(20) DEFAULT NULL COMMENT '网络设备角色ID',
    `cabinet_id`             bigint(20) DEFAULT NULL COMMENT '机柜id',
    `machine_room_abbr`      varchar(255) DEFAULT NULL COMMENT '机房缩写',
    `machine_room_number`    varchar(255) DEFAULT NULL COMMENT '机房编号',
    `cabinet_number`         varchar(255) DEFAULT NULL COMMENT '机柜编号',
    `slot_position`          varchar(255) DEFAULT NULL COMMENT '槽位',
    `u_number`               int          DEFAULT NULL COMMENT 'U数',
    `create_user_id`         varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`            datetime NULL DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品基线';


CREATE TABLE `server_shelve`
(
    `id`                      bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `plan_id`                 bigint(20) DEFAULT NULL COMMENT '方案id',
    `sort_number`             int          DEFAULT NULL COMMENT '序号',
    `node_role_id`            bigint(20) DEFAULT NULL COMMENT '节点角色id',
    `sn`                      varchar(255) DEFAULT NULL COMMENT 'SN',
    `model`                   varchar(255) DEFAULT NULL COMMENT '机型',
    `cabinet_id`              bigint(20) DEFAULT NULL COMMENT '机柜id',
    `machine_room_abbr`       varchar(255) DEFAULT NULL COMMENT '机房缩写',
    `machine_room_number`     varchar(255) DEFAULT NULL COMMENT '机房编号',
    `column_number`           varchar(255) DEFAULT NULL COMMENT '列号',
    `cabinet_asw`             varchar(255) DEFAULT NULL COMMENT '机柜ASW组',
    `cabinet_number`          varchar(255) DEFAULT NULL COMMENT '机柜编号',
    `cabinet_original_number` varchar(255) DEFAULT NULL COMMENT '机柜原始编号',
    `cabinet_location`        varchar(255) DEFAULT NULL COMMENT '机柜位置',
    `slot_position`           varchar(255) DEFAULT NULL COMMENT '槽位',
    `network_interface`       varchar(255) DEFAULT NULL COMMENT '网络接口',
    `bmc_user_name`           varchar(255) DEFAULT NULL COMMENT 'bmc用户名',
    `bmc_password`            varchar(255) DEFAULT NULL COMMENT 'bmc密码',
    `bmc_ip`                  varchar(255) DEFAULT NULL COMMENT 'bmc IP地址',
    `bmc_mac`                 varchar(255) DEFAULT NULL COMMENT 'bmc mac地址',
    `mask`                    varchar(255) DEFAULT NULL COMMENT '掩码',
    `gateway`                 varchar(255) DEFAULT NULL COMMENT '网关',
    `create_user_id`          varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`             datetime NULL DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='server_shelve';


CREATE TABLE `cabinet_info`
(
    `id`                       bigint(20)   NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`                  bigint(20)   DEFAULT NULL COMMENT '方案id',
    `machine_room_abbr`        varchar(100) DEFAULT NULL COMMENT '机房缩写',
    `machine_room_num`         varchar(50)  DEFAULT NULL COMMENT '房间号',
    `column_num`               varchar(50)  DEFAULT NULL COMMENT '列号',
    `cabinet_num`              varchar(50)  DEFAULT NULL COMMENT '机柜编号',
    `original_num`             varchar(50)  DEFAULT NULL COMMENT '原始编号',
    `cabinet_type`             tinyint(4)   DEFAULT NULL COMMENT '机柜类型，1：网络机柜，2：服务机柜，3：存储机柜',
    `business_attribute`       varchar(255) DEFAULT NULL COMMENT '业务属性',
    `cabinet_asw`              varchar(255) DEFAULT NULL COMMENT '机柜ASW组',
    `total_power`              int(11)      DEFAULT NULL COMMENT '总功率（W）',
    `residual_power`           int(11)      DEFAULT NULL COMMENT '剩余功率（W）',
    `total_slot_num`           int(11)      DEFAULT NULL COMMENT '总槽位数（U位）',
    `idle_slot_range`          varchar(255) DEFAULT NULL COMMENT '空闲槽位（U位）范围',
    `max_rack_server_num`      int(11)      DEFAULT NULL COMMENT '最大可上架服务器数',
    `residual_rack_server_num` int(11)      DEFAULT NULL COMMENT '剩余上架服务器数',
    `rack_server_slot`         varchar(255) DEFAULT NULL COMMENT '已上架服务器（U位）',
    `residual_rack_asw_port`   varchar(255) DEFAULT NULL COMMENT '剩余可上架ASW端口',
    `create_time`              datetime     DEFAULT NULL COMMENT '创建时间',
    `update_time`              datetime     DEFAULT NULL COMMENT '修改时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='机柜信息';


CREATE TABLE `cabinet_idle_slot_rel`
(
    `cabinet_id`       bigint(20) DEFAULT NULL COMMENT '机柜id',
    `idle_slot_number` int(11) DEFAULT NULL COMMENT '空闲槽位（U位）号'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机柜空闲槽位关联表';


CREATE TABLE `cabinet_rack_server_slot_rel`
(
    `cabinet_id`           bigint(20) DEFAULT NULL COMMENT '机柜id',
    `rack_server_slot_num` int(11) DEFAULT NULL COMMENT '已上架服务器槽位（U位）号'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机柜已上架服务器槽位关联表';


CREATE TABLE `cabinet_rack_asw_port_rel`
(
    `cabinet_id`                 bigint(20) DEFAULT NULL COMMENT '机柜id',
    `residual_rack_asw_port_num` int(11) DEFAULT NULL COMMENT '剩余可上架ASW端口号'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机柜剩余可上架ASW端口关联表';


CREATE TABLE `vlan_id_config`
(
    `id`                    bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`               bigint(20) DEFAULT NULL COMMENT '方案id',
    `in_band_mgt_vlan_id`   varchar(100) DEFAULT NULL COMMENT '带内管理Vlan ID',
    `local_storage_vlan_id` varchar(100) DEFAULT NULL COMMENT '本地存储网Vlan ID',
    `biz_intranet_vlan_id`  varchar(100) DEFAULT NULL COMMENT '业务内网Vlan ID',
    `create_user_id`        varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`           datetime     DEFAULT NULL COMMENT '创建时间',
    `update_user_id`        varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`           datetime     DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='Vlan ID配置表';


CREATE TABLE `cell_config`
(
    `id`                           bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`                      bigint(20) DEFAULT NULL COMMENT '方案id',
    `biz_region_abbr`              varchar(100) DEFAULT NULL COMMENT '业务区域缩写',
    `cell_self_mgt`                tinyint(4) DEFAULT NULL COMMENT '集群自纳管，0：否，1：是',
    `mgt_global_dns_root_domain`   varchar(255) DEFAULT NULL COMMENT '管理网全局DNS根域',
    `global_dns_svc_address`       varchar(255) DEFAULT NULL COMMENT '全局DNS服务地址',
    `cell_vip`                     varchar(255) DEFAULT NULL COMMENT '集群Vip',
    `cell_vip_ipv6`                varchar(255) DEFAULT NULL COMMENT '集群Vip-IPV6地址',
    `external_ntp_ip`              varchar(500) DEFAULT NULL COMMENT '外部时钟源IP（多个时钟源以逗号分隔）',
    `network_mode`                 tinyint(4) DEFAULT NULL COMMENT '组网模式，0：标准模式，1：纯二层组网模式',
    `cell_container_network`       varchar(255) DEFAULT NULL COMMENT '集群网络配置-集群容器网',
    `cell_container_network_ipv6`  varchar(255) DEFAULT NULL COMMENT '集群网络配置-集群容器网IPV6',
    `cell_svc_network`             varchar(255) DEFAULT NULL COMMENT '集群网络配置-集群服务网',
    `cell_svc_network_ipv6`        varchar(255) DEFAULT NULL COMMENT '集群网络配置-集群服务网IPV6',
    `add_cell_node_ssh_public_key` varchar(255) DEFAULT NULL COMMENT '添加集群节点SSH访问公钥',
    `create_user_id`               varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`                  datetime     DEFAULT NULL COMMENT '创建时间',
    `update_user_id`               varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`                  datetime     DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='集群配置表';


CREATE TABLE `route_planning_config`
(
    `id`                              bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`                         bigint(20) DEFAULT NULL COMMENT '方案id',
    `deploy_use_bgp`                  tinyint(4) DEFAULT NULL COMMENT '使用BGP部署，0：否，1：是',
    `deploy_mach_switch_self_num`     varchar(255) DEFAULT NULL COMMENT '部署机所在交换机自治号',
    `deploy_mach_switch_ip`           varchar(500) DEFAULT NULL COMMENT '部署机所在交换机IP（多个IP以逗号分隔）',
    `svc_external_access_address`     varchar(500) DEFAULT NULL COMMENT '服务外部访问地址',
    `bgp_neighbor`                    varchar(255) DEFAULT NULL COMMENT 'BGP邻居',
    `cell_dns_svc_address`            varchar(255) DEFAULT NULL COMMENT '集群DNS服务地址',
    `region_dns_svc_address`          varchar(255) DEFAULT NULL COMMENT 'Region DNS服务地址',
    `ops_center_ip`                   varchar(255) DEFAULT NULL COMMENT '运维中心访问IP',
    `ops_center_ipv6`                 varchar(255) DEFAULT NULL COMMENT '运维中心访问IPV6地址',
    `ops_center_port`                 varchar(255) DEFAULT NULL COMMENT '运维中心访问端口',
    `ops_center_domain`               varchar(255) DEFAULT NULL COMMENT '运维中心访问域名',
    `operation_center_ip`             varchar(255) DEFAULT NULL COMMENT '运营中心访问IP',
    `operation_center_ipv6`           varchar(255) DEFAULT NULL COMMENT '运营中心访问IPV6地址',
    `operation_center_port`           varchar(255) DEFAULT NULL COMMENT '运营中心访问端口',
    `operation_center_domain`         varchar(255) DEFAULT NULL COMMENT '运营中心访问域名',
    `ops_center_init_user_name`       varchar(255) DEFAULT NULL COMMENT '运维中心初始化用户配置-用户名',
    `ops_center_init_user_pwd`        varchar(255) DEFAULT NULL COMMENT '运维中心初始化用户配置-密码',
    `operation_center_init_user_name` varchar(255) DEFAULT NULL COMMENT '运营中心初始化用户配置-用户名',
    `operation_center_init_user_pwd`  varchar(255) DEFAULT NULL COMMENT '运营中心初始化用户配置-密码',
    `create_user_id`                  varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`                     datetime     DEFAULT NULL COMMENT '创建时间',
    `update_user_id`                  varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`                     datetime     DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='路由规划配置表';


CREATE TABLE `large_network_segment_config`
(
    `id`                                 bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`                            bigint(20) DEFAULT NULL COMMENT '方案id',
    `storage_network_segment_route`      varchar(500) DEFAULT NULL COMMENT '存储前端网规划网段明细路由',
    `biz_intranet_network_segment_route` varchar(500) DEFAULT NULL COMMENT '业务内网规划网段明细路由',
    `biz_external_large_network_segment` varchar(500) DEFAULT NULL COMMENT '业务外网大网网段',
    `bmc_network_segment_route`          varchar(500) DEFAULT NULL COMMENT 'bmc规划网段明细路由',
    `create_user_id`                     varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`                        datetime     DEFAULT NULL COMMENT '创建时间',
    `update_user_id`                     varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`                        datetime     DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='大网网段配置表';


CREATE TABLE `ip_demand_shelve`
(
    `id`               BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `plan_id`          bigint(20) NOT NULL COMMENT '方案ID',
    `logical_grouping` varchar(255) DEFAULT NULL COMMENT '逻辑分组',
    `segment_type`     varchar(255) DEFAULT NULL COMMENT '网段类型',
    `network_type`     tinyint(4) DEFAULT NULL COMMENT '网络类型，0：ipv4，1：ipv6',
    `vlan`             varchar(45)  DEFAULT NULL COMMENT 'VLAN ID',
    `c_num`            varchar(45)  DEFAULT NULL COMMENT 'C数量',
    `address`          varchar(255) DEFAULT NULL COMMENT '地址段',
    `describe`         varchar(255) DEFAULT NULL COMMENT '描述',
    `address_planning` varchar(255) DEFAULT NULL COMMENT 'IP地址规划建议',
    `create_user_id`   varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`      datetime NULL DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='Ip需求规划表';


CREATE TABLE `network_device_ip`
(
    `id`                            bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`                       bigint(20) DEFAULT NULL COMMENT '方案id',
    `logical_grouping`              varchar(255) DEFAULT NULL COMMENT '网络设备逻辑分组',
    `pxe_subnet`                    varchar(100) DEFAULT NULL COMMENT 'pxe子网',
    `pxe_subnet_range`              varchar(1000) DEFAULT NULL COMMENT 'pxe子网范围',
    `pxe_network_gateway`           varchar(100) DEFAULT NULL COMMENT 'pxe网网关',
    `manage_subnet`                 varchar(100) DEFAULT NULL COMMENT '管理网子网',
    `manage_network_gateway`        varchar(100) DEFAULT NULL COMMENT '管理网网关',
    `manage_ipv6_subnet`            varchar(100) DEFAULT NULL COMMENT '管理网IPV6子网',
    `manage_ipv6_network_gateway`   varchar(100) DEFAULT NULL COMMENT '管理网IPV6网关',
    `biz_subnet`                    varchar(100) DEFAULT NULL COMMENT '业务网子网',
    `biz_network_gateway`           varchar(100) DEFAULT NULL COMMENT '业务网网关',
    `storage_front_network`         varchar(100) DEFAULT NULL COMMENT '存储前端网',
    `storage_front_network_gateway` varchar(100) DEFAULT NULL COMMENT '存储前端网网关',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='网络设备ip分配表';


CREATE TABLE `server_ip`
(
    `id`                  bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`             bigint(20) DEFAULT NULL COMMENT '方案id',
    `sn`                  varchar(255) DEFAULT NULL COMMENT 'SN',
    `manage_network_ip`   varchar(100) DEFAULT NULL COMMENT '管理网ip',
    `manage_network_ipv6` varchar(100) DEFAULT NULL COMMENT '管理网ipv6',
    `biz_intranet_ip`     varchar(100) DEFAULT NULL COMMENT '业务内网ip',
    `storage_network_ip`  varchar(100) DEFAULT NULL COMMENT '存储网ip',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器ip分配表';


CREATE TABLE `software_bom_license_baseline`
(
    `id`                  bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`          bigint(20) DEFAULT NULL COMMENT '版本id',
    `cloud_service`       varchar(255) DEFAULT NULL COMMENT '云服务',
    `service_code`        varchar(255) DEFAULT NULL COMMENT '服务编码',
    `sell_specs`          varchar(255) DEFAULT NULL COMMENT '售卖规格',
    `value_added_service` varchar(255) DEFAULT NULL COMMENT '增值服务',
    `authorized_unit`     varchar(255) DEFAULT NULL COMMENT '授权单元',
    `sell_type`           varchar(255) DEFAULT NULL COMMENT '售卖类型',
    `hardware_arch`       varchar(255) DEFAULT NULL COMMENT '硬件架构',
    `bom_id`              varchar(255) DEFAULT NULL COMMENT 'bom id',
    `calc_method`         varchar(255) DEFAULT NULL COMMENT '计算方式',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT='软件BOM/License基线表';


CREATE TABLE `software_bom_planning`
(
    `id`                   bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `plan_id`              bigint(20) DEFAULT NULL COMMENT '方案id',
    `software_baseline_id` bigint(20) DEFAULT NULL COMMENT '软件基线id',
    `bom_id`               varchar(255) DEFAULT NULL COMMENT 'bom id',
    `number`               int          DEFAULT NULL COMMENT '数量',
    `cloud_service`        varchar(255) DEFAULT NULL COMMENT '云服务',
    `service_code`         varchar(255) DEFAULT NULL COMMENT '服务编码',
    `sell_specs`           varchar(255) DEFAULT NULL COMMENT '售卖规格',
    `authorized_unit`      varchar(255) DEFAULT NULL COMMENT '授权单元',
    `sell_type`            varchar(255) DEFAULT NULL COMMENT '售卖类型',
    `hardware_arch`        varchar(255) DEFAULT NULL COMMENT '硬件架构',
    `value_added_service`  varchar(255) DEFAULT NULL COMMENT '增值服务',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品规划bom表';

CREATE TABLE IF NOT EXISTS `feature_name_code_rel`
(
    `id`                    int NOT NULL AUTO_INCREMENT COMMENT 'id',
    `feature_name`          varchar(255) DEFAULT NULL COMMENT 'feature名称',
    `feature_code`          varchar(255) DEFAULT NULL COMMENT 'feature_code',
    `feature_type`          varchar(255) DEFAULT NULL COMMENT '类别：云产品、服务器、网络',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='feature名称与code对应表';

insert into `feature_name_code_rel` (`feature_name`, `feature_code`, `feature_type`)
values 	("运营运维", "OM", "cloud_product"),
          ("计算服务", "Compute", "cloud_product"),
          ("存储服务", "Storage", "cloud_product"),
          ("网络服务", "Network", "cloud_product"),
          ("安全服务", "Security", "cloud_product"),
          ("大数据", "Bigdata", "cloud_product"),
          ("数据库", "Database", "cloud_product"),
          ("应用服务", "AppService", "cloud_product"),
          ("应用商店", "AppMarket", "cloud_product"),
          ("平台规模授权", "Platform", "cloud_product"),
          ("软件Base", "SoftwareBase", "cloud_product"),
          ("平台升级维保","ServiceYear","cloud_product"),

          ("管理区", "master_node", "server_node"),
          ("存储区", "storage_node", "server_node"),
          ("业务区", "worker_node", "server_node"),
          ("网络区", "network_node", "server_node");
