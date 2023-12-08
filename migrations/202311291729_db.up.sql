CREATE TABLE `customer_manage`
(
    `id`             bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id编码',
    `customer_name`  varchar(50)  DEFAULT NULL COMMENT '客户名称',
    `leader_id`      varchar(50)  DEFAULT NULL COMMENT '客户接口人id',
    `leader_name`    varchar(255) DEFAULT NULL COMMENT '客户接口人名称',
    `create_user_id` varchar(50)  DEFAULT NULL COMMENT '创建人id',
    `create_time`    datetime     DEFAULT NULL COMMENT '创建时间',
    `update_user_id` varchar(50)  DEFAULT NULL COMMENT '更新人id',
    `update_time`    datetime     DEFAULT NULL COMMENT '更新时间',
    `delete_state`   tinyint(1) DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='客户表';


CREATE TABLE `user_manage`
(
    `id`                varchar(50) NOT NULL COMMENT '用户id（ldap uid）',
    `user_name`         varchar(255) DEFAULT NULL COMMENT '用户名',
    `employee_number`   varchar(20)  DEFAULT NULL COMMENT '工号',
    `telephone_number`  varchar(20)  DEFAULT NULL COMMENT '电话号码',
    `department`        varchar(255) DEFAULT NULL COMMENT '部门',
    `office_name`       varchar(255) DEFAULT NULL COMMENT '办公部门名',
    `department_number` varchar(20)  DEFAULT NULL COMMENT '部门id',
    `mail`              varchar(255) DEFAULT NULL COMMENT '邮箱',
    `delete_state`      tinyint(1) DEFAULT 0 COMMENT '作废状态：1，作废；0，正常。',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='用户表';


CREATE TABLE `role_manage`
(
    `user_id` varchar(50) CHARACTER SET utf8mb4 NOT NULL COMMENT '用户id',
    `role`    varchar(20)                       NOT NULL COMMENT '角色权限：admin：超级管理员；normal：普通用户',
    PRIMARY KEY (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT='用户角色关联表';


CREATE TABLE `permissions_manage`
(
    `user_id`      varchar(50)  NOT NULL DEFAULT '0' COMMENT '用户id(成员id)',
    `user_name`    varchar(255) NOT NULL DEFAULT '0' COMMENT '成员名称',
    `customer_id`  bigint(20) NOT NULL COMMENT '客户id',
    `delete_state` tinyint(1) NOT NULL DEFAULT 0 COMMENT '作废状态：0，正常；1，作废'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=COMPACT COMMENT='项目成员关联表';


CREATE TABLE `cloud_platform_manage`
(
    `id`             bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `name`           varchar(255) DEFAULT NULL COMMENT '云平台名称',
    `type`           varchar(50)  DEFAULT NULL COMMENT '云平台类型',
    `customer_id`    bigint(20) DEFAULT NULL COMMENT '客户id',
    `create_user_id` varchar(255) DEFAULT NULL COMMENT '创建用户',
    `create_time`    datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id` varchar(255) DEFAULT NULL COMMENT '更新人',
    `update_time`    datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`   tinyint(1) DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='云平台';


CREATE TABLE `region_manage`
(
    `id`                bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'regionId',
    `code`              varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT 'region编码',
    `name`              varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT 'region名称',
    `type`              varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT 'region类型',
    `cloud_platform_id` bigint(20) NULL DEFAULT NULL COMMENT '云平台id',
    `create_user_id`    varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '创建人id',
    `create_time`       datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id`    varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '更新人id',
    `update_time`       datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`      tinyint(1) NULL DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;


CREATE TABLE `az_manage`
(
    `id`                bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'azId',
    `code`              varchar(255) DEFAULT NULL COMMENT 'az编码',
    `region_id`         bigint(20) NULL DEFAULT NULL COMMENT 'regionId',
    `machine_room_name` varchar(255) NULL DEFAULT NULL COMMENT '机房名称',
    `machine_room_code` varchar(255) NULL DEFAULT NULL COMMENT '机房编码',
    `province`          varchar(50) NULL DEFAULT NULL COMMENT '省',
    `city`              varchar(50) NULL DEFAULT NULL COMMENT '市',
    `address`           varchar(50) NULL DEFAULT NULL COMMENT '地址',
    `create_user_id`    varchar(255) NULL DEFAULT NULL COMMENT '创建人id',
    `create_time`       datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id`    varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`       datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`      tinyint(1) NULL DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = 'az管理表' ROW_FORMAT = Dynamic;


CREATE TABLE `cell_manage`
(
    `id`             bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'cell Id',
    `name`           varchar(255) DEFAULT NULL COMMENT 'cell名称',
    `create_user_id` varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`    datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id` varchar(255) NULL DEFAULT NULL COMMENT '更新人id',
    `update_time`    datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`   tinyint(1) NULL DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = 'az管理表' ROW_FORMAT = Dynamic;


CREATE TABLE `project_manage`
(
    `id`                bigint(20) NOT NULL AUTO_INCREMENT COMMENT '项目id',
    `name`              varchar(255) DEFAULT NULL COMMENT '项目名称',
    `cloud_platform_id` bigint(20) NULL DEFAULT NULL COMMENT '云平台id',
    `region_id`         bigint(20) NULL DEFAULT NULL COMMENT 'regionId',
    `az_id`             bigint(20) NULL DEFAULT NULL COMMENT 'azId',
    `cell_id`           bigint(20) NULL DEFAULT NULL COMMENT 'cell Id',
    `customer_id`       bigint(20) NULL DEFAULT NULL COMMENT '客户id',
    `type`              varchar(50)  DEFAULT NULL COMMENT '项目类型',
    `stage`             varchar(50)  DEFAULT NULL COMMENT '项目阶段',
    `create_user_id`    varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`       datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id`    varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`       datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`      tinyint(1) NULL DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '项目管理表' ROW_FORMAT = Dynamic;


CREATE TABLE `plan_manage`
(
    `id`                  bigint(20) NOT NULL AUTO_INCREMENT COMMENT '方案id',
    `name`                varchar(255) DEFAULT NULL COMMENT '方案名称',
    `type`                varchar(50)  DEFAULT NULL COMMENT '方案类型',
    `stage`               varchar(50)  DEFAULT NULL COMMENT '方案阶段',
    `project_id`          int NULL DEFAULT NULL COMMENT '项目id',
    `create_user_id`      varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`         datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id`      varchar(255) DEFAULT NULL COMMENT '更新人id',
    `update_time`         datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`        tinyint(1) NULL DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    `business_plan_stage` tinyint(1) NULL DEFAULT NULL COMMENT '业务规划阶段：0，业务规划开始阶段；1，云产品配置阶段；2，服务器规划阶段；3，网络设备规划阶段； 4，业务规划结束',
    `deliver_plan_stage`  tinyint(1) NULL DEFAULT NULL COMMENT '交付规划阶段：0，交付规划开始阶段',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;


CREATE TABLE `config_item`
(
    `id`   int NULL DEFAULT NULL COMMENT '配置Id',
    `p_id` int NULL DEFAULT NULL COMMENT '上级配置Id',
    `name` varchar(255) DEFAULT NULL COMMENT '配置名称',
    `code` varchar(255) DEFAULT NULL COMMENT '配置编码',
    `data` varchar(255) DEFAULT NULL COMMENT '配置值',
    `sort` int NULL DEFAULT NULL COMMENT '排序'
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '配置表' ROW_FORMAT = Dynamic;

INSERT INTO `config_item` VALUES (1, 0, '云平台类型', 'cloudPlatformType', NULL, NULL);
INSERT INTO `config_item` VALUES (101, 1, '运营云', 'operational', NULL, 1);
INSERT INTO `config_item` VALUES (102, 1, '交付云', 'delivery', NULL, 2);
INSERT INTO `config_item` VALUES (2, 0, '区域类型', 'regionType', NULL, NULL);
INSERT INTO `config_item` VALUES (201, 2, '合设区域', 'merge', NULL, 1);
INSERT INTO `config_item` VALUES (202, 2, '业务区域', 'business', NULL, 2);
INSERT INTO `config_item` VALUES (203, 2, '管理区域', 'manage', NULL, 3);
INSERT INTO `config_item` VALUES (3, 0, '项目类型', 'projectType', NULL, NULL);
INSERT INTO `config_item` VALUES (301, 3, '新建', 'create', NULL, 1);
INSERT INTO `config_item` VALUES (302, 3, '扩容', 'expansion', NULL, 2);
INSERT INTO `config_item` VALUES (303, 3, '升级', 'upgradation', NULL, 3);
INSERT INTO `config_item` VALUES (4, 0, '项目阶段', 'projectStage', NULL, NULL);
INSERT INTO `config_item` VALUES (401, 4, '规划阶段', 'planning', NULL, 1);
INSERT INTO `config_item` VALUES (402, 4, '交付阶段', 'delivery', NULL, 2);
INSERT INTO `config_item` VALUES (403, 4, '已交付', 'delivered', NULL, 3);
INSERT INTO `config_item` VALUES (5, 0, '方案类型', 'planType', NULL, NULL);
INSERT INTO `config_item` VALUES (501, 5, '普通方案', 'general', NULL, 1);
INSERT INTO `config_item` VALUES (502, 5, '备用方案', 'standby', NULL, 2);
INSERT INTO `config_item` VALUES (503, 5, '交付方案', 'delivery', NULL, 3);
INSERT INTO `config_item` VALUES (6, 0, '方案阶段', 'planStage', NULL, NULL);
INSERT INTO `config_item` VALUES (601, 6, '待规划', 'plan', NULL, 1);
INSERT INTO `config_item` VALUES (602, 6, '规划中', 'planning', NULL, 2);
INSERT INTO `config_item` VALUES (603, 6, '规划完成', 'planned', NULL, 3);
INSERT INTO `config_item` VALUES (604, 6, '交付中', 'delivering', NULL, 4);
INSERT INTO `config_item` VALUES (605, 6, '交付完成', 'delivered', NULL, 5);


CREATE TABLE `cloud_product_baseline`
(
    `id`               bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`       bigint(20) DEFAULT NULL COMMENT '软件版本id',
    `product_type`     varchar(100) DEFAULT NULL COMMENT '产品类型',
    `product_name`     varchar(255) DEFAULT NULL COMMENT '产品名称',
    `product_code`     varchar(100) DEFAULT NULL COMMENT '产品code',
    `sell_specs`       varchar(255) DEFAULT NULL COMMENT '售卖规格',
    `authorized_unit`  varchar(255) DEFAULT NULL COMMENT '授权单元',
    `whether_required` tinyint(4) DEFAULT NULL COMMENT '是否必选，0：否，1：是',
    `instructions`     varchar(500) DEFAULT NULL COMMENT '说明',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品基线';


CREATE TABLE `server_planning`
(
    `id`                 bigint(20) NOT NULL AUTO_INCREMENT COMMENT '服务器规划id',
    `plan_id`            bigint(20) DEFAULT NULL COMMENT ' 方案id',
    `node_role_id`       bigint(20) DEFAULT NULL COMMENT '节点角色id',
    `mixed_node_role_id` bigint(20) DEFAULT NULL COMMENT '节点角色id',
    `server_baseline_id` bigint(20) DEFAULT NULL COMMENT '服务器基线表id',
    `number`             int          DEFAULT NULL COMMENT '数量',
    `open_dpdk`          int          DEFAULT NULL COMMENT '是否开启DPDK，1：开启，0：关闭',
    `create_user_id`     varchar(255) DEFAULT NULL COMMENT '创建人id',
    `create_time`        datetime NULL DEFAULT NULL COMMENT '创建时间',
    `update_user_id`     varchar(255) NULL DEFAULT NULL COMMENT '更新人id',
    `update_time`        datetime NULL DEFAULT NULL COMMENT '更新时间',
    `delete_state`       tinyint(1) NULL DEFAULT NULL COMMENT '作废状态：1，作废；0，正常',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品基线';


CREATE TABLE `software_version`
(
    `id`                  bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `software_version`    varchar(255) DEFAULT NULL COMMENT '版本号',
    `cloud_platform_type` varchar(100) DEFAULT NULL COMMENT '云平台类型',
    `release_time`        datetime     DEFAULT NULL COMMENT '发布时间',
    `create_time`         datetime     DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='软件版本表';


CREATE TABLE `node_role_baseline`
(
    `id`             bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`     bigint(20) DEFAULT NULL COMMENT '软件版本id',
    `node_role_code` varchar(255) DEFAULT NULL COMMENT '节点角色编码',
    `node_role_name` varchar(255) DEFAULT NULL COMMENT '节点角色名称',
    `minimum_num`    int(11) DEFAULT NULL COMMENT '单独部署最小数量',
    `deploy_method`  varchar(255) DEFAULT NULL COMMENT '部署方式',
    `support_dpdk`   tinyint(4) DEFAULT NULL COMMENT '是否支持DPDK，0：否，1：是',
    `classify`       varchar(255) DEFAULT NULL COMMENT '分类',
    `annotation`     varchar(500) DEFAULT NULL COMMENT '节点说明',
    `business_type`  varchar(255) DEFAULT NULL COMMENT '业务类型',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点角色基线表';


CREATE TABLE `node_role_mixed_deploy`
(
    `node_role_id`       bigint(20) NOT NULL COMMENT '节点角色id',
    `mixed_node_role_id` bigint(20) NOT NULL COMMENT '混部的节点角色id'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点角色混部表';


CREATE TABLE `network_device_list`
(
    `id`                       bigint(20) NOT NULL AUTO_INCREMENT,
    `plan_id`                  bigint(20) NOT NULL COMMENT '方案ID',
    `network_device_role`      varchar(45)   DEFAULT NULL COMMENT '设备类型->网络设备角色编码',
    `network_device_role_name` varchar(255)  DEFAULT NULL COMMENT '设备类型->网络设备角色名称',
    `network_device_role_id`   bigint(20) DEFAULT NULL COMMENT '设备类型->网络设备角色ID',
    `logical_grouping`         varchar(255)  DEFAULT NULL COMMENT '逻辑分组',
    `device_id`                varchar(255)  DEFAULT NULL COMMENT '设备ID',
    `conf_overview`            varchar(1000) DEFAULT NULL COMMENT '配置概述',
    `brand`                    varchar(45)   DEFAULT NULL COMMENT '厂商',
    `device_model`             varchar(45)   DEFAULT NULL COMMENT '设备型号',
    `create_time`              datetime      DEFAULT NULL COMMENT '创建时间',
    `update_time`              datetime      DEFAULT NULL COMMENT '修改时间',
    `delete_state`             tinyint(4) DEFAULT NULL COMMENT '删除状态0：未删除；1：已删除',
    PRIMARY KEY (`id`) USING BTREE,
    KEY                        `IDX_U_PLAN_ID_STATE` (`plan_id`,`delete_state`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='网络设备清单';


CREATE TABLE `network_device_planning`
(
    `id`                     bigint(20) NOT NULL AUTO_INCREMENT,
    `plan_id`                bigint(20) NOT NULL COMMENT '方案ID',
    `brand`                  varchar(45) CHARACTER SET utf8 DEFAULT NULL COMMENT '厂商',
    `application_dispersion` char(1)                        DEFAULT '1' COMMENT '应用分散度: 1-分散在不同服务器',
    `aws_server_num`         tinyint(2) DEFAULT NULL COMMENT 'AWS下连服务器数44/45',
    `aws_box_num`            tinyint(2) DEFAULT NULL COMMENT '每组AWS几个机柜4/3',
    `total_box_num`          tinyint(4) DEFAULT NULL COMMENT '机柜估算数量',
    `create_time`            datetime                       DEFAULT NULL COMMENT '创建时间',
    `update_time`            datetime                       DEFAULT NULL COMMENT '更新时间',
    `ipv6`                   char(1)                        DEFAULT '0' COMMENT '是否为ipv4/ipv6双栈交付 0：ipv4交付 1:ipv4/ipv6双栈交付',
    `network_model`          tinyint(4) DEFAULT 1 COMMENT '组网模型: 1-三网合一  2-两网分离  3-三网分离',
    `device_type`            tinyint(4) DEFAULT NULL COMMENT '设备类型，0：信创，1：商用',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `IDX_U_PLAN_ID` (`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='网络设备规划表';


CREATE TABLE `server_baseline`
(
    `id`                    bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`            bigint(20) DEFAULT NULL COMMENT '版本id',
    `arch`                  varchar(50)   DEFAULT NULL COMMENT '硬件架构',
    `network_interface`     varchar(50)   DEFAULT NULL COMMENT '网络接口',
    `bom_code`              varchar(255)  DEFAULT NULL COMMENT 'BOM编码',
    `configuration_info`    varchar(1000) DEFAULT NULL COMMENT '配置概要',
    `spec`                  varchar(255)  DEFAULT NULL COMMENT '规格',
    `cpu_type`              varchar(50)   DEFAULT NULL COMMENT 'CPU类型',
    `cpu`                   int(11) DEFAULT NULL COMMENT 'CPU核数',
    `gpu`                   varchar(255)  DEFAULT NULL COMMENT 'GPU',
    `memory`                int(11) DEFAULT NULL COMMENT '内存',
    `system_disk_type`      varchar(20)   DEFAULT NULL COMMENT '系统盘类型',
    `system_disk`           varchar(255)  DEFAULT NULL COMMENT '系统盘',
    `storage_disk_type`     varchar(50)   DEFAULT NULL COMMENT '存储盘类型',
    `storage_disk_num`      int(11) DEFAULT NULL COMMENT '存储盘数量',
    `storage_disk_capacity` int(11) DEFAULT NULL COMMENT '存储盘单盘容量（G）',
    `ram_disk`              varchar(255)  DEFAULT NULL COMMENT '缓存盘',
    `network_card_num`      int(11) DEFAULT NULL COMMENT '网卡数量',
    `power`                 int(11) DEFAULT NULL COMMENT '功率',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器基线表';


CREATE TABLE `server_node_role_rel`
(
    `server_id`    bigint(20) DEFAULT NULL COMMENT '服务器id',
    `node_role_id` bigint(20) DEFAULT NULL COMMENT '节点角色id'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器与节点角色关联表';


CREATE TABLE `network_device_role_baseline`
(
    `id`                bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`        bigint(20) DEFAULT NULL COMMENT '版本id',
    `device_type`       varchar(255) DEFAULT NULL COMMENT '设备类型',
    `func_type`         varchar(255) DEFAULT NULL COMMENT '类型',
    `func_compo_name`   varchar(255) DEFAULT NULL COMMENT '功能组件',
    `func_compo_code`   varchar(255) DEFAULT NULL COMMENT '功能组件命名',
    `description`       varchar(500) DEFAULT NULL COMMENT '描述',
    `two_network_iso`   tinyint(4) DEFAULT NULL COMMENT '两网分离，0：否，1：是，2：需要查询网络模式与节点角色或者网络设备角色关联表',
    `three_network_iso` tinyint(4) DEFAULT NULL COMMENT '三网分离，0：否，1：是，2：需要查询网络模式与节点角色或者网络设备角色关联表',
    `triple_play`       tinyint(4) DEFAULT NULL COMMENT '三网合一，0：否，1：是，2：需要查询网络模式与节点角色或者网络设备角色关联表',
    `minimum_num_unit`  int(11) DEFAULT NULL COMMENT '最小单元数',
    `unit_device_num`   int(11) DEFAULT NULL COMMENT '单元设备数量',
    `design_spec`       varchar(500) DEFAULT NULL COMMENT '设计规格',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网络设备角色表';


CREATE TABLE `network_device_baseline`
(
    `id`            bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`    bigint(20) DEFAULT NULL COMMENT '版本id',
    `device_model`  varchar(255)  DEFAULT NULL COMMENT '设备型号',
    `manufacturer`  varchar(255)  DEFAULT NULL COMMENT '厂商',
    `device_type`   tinyint(4) DEFAULT NULL COMMENT '信创/商用， 0：信创，1：商用',
    `network_model` varchar(255)  DEFAULT NULL COMMENT '网络模型',
    `conf_overview` varchar(1000) DEFAULT NULL COMMENT '配置概述',
    `purpose`       varchar(500)  DEFAULT NULL COMMENT '用途',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网络设备基线表';


CREATE TABLE `network_device_role_rel`
(
    `device_id`      bigint(20) DEFAULT NULL COMMENT '设备id',
    `device_role_id` bigint(20) DEFAULT NULL COMMENT '设备角色id'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网络设备与设备角色关联表';


CREATE TABLE `ip_demand_baseline`
(
    `id`            bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `version_id`    bigint(20) DEFAULT NULL COMMENT '版本id',
    `vlan`          varchar(20)  DEFAULT NULL COMMENT 'vlan id',
    `explain`       varchar(500) DEFAULT NULL COMMENT '说明',
    `description`   varchar(500) DEFAULT NULL COMMENT '描述',
    `ip_suggestion` varchar(500) DEFAULT NULL COMMENT 'IP地址规划建议',
    `assign_num`    varchar(100) DEFAULT NULL COMMENT '分配数量',
    `remark`        varchar(500) DEFAULT NULL COMMENT '备注',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP需求规划基线表';


CREATE TABLE `ip_demand_device_role_rel`
(
    `ip_demand_id`   bigint(20) DEFAULT NULL COMMENT 'IP需求规划id',
    `device_role_id` bigint(20) DEFAULT NULL COMMENT '网络设备角色id'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IP需求规划与网络设备角色关联表';


CREATE TABLE `cloud_product_depend_rel`
(
    `product_id`        bigint(20) DEFAULT NULL COMMENT '云产品id',
    `depend_product_id` bigint(20) DEFAULT NULL COMMENT '依赖的云产品id'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品依赖关系表';


CREATE TABLE `cloud_product_node_role_rel`
(
    `product_id`     bigint(20) DEFAULT NULL COMMENT '云产品id',
    `node_role_id`   bigint(20) DEFAULT NULL COMMENT '节点角色id',
    `node_role_type` tinyint(4) DEFAULT NULL COMMENT '节点角色类型，1：管控资源节点角色，0：资源节点角色'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品与节点角色关联表';


CREATE TABLE `cloud_product_planning`
(
    `id`           bigint(20) NOT NULL AUTO_INCREMENT COMMENT '配置id',
    `plan_id`      bigint(20) DEFAULT NULL COMMENT '方案id',
    `product_id`   bigint(20) DEFAULT NULL COMMENT '云产品id',
    `sell_spec`    varchar(60) DEFAULT NULL COMMENT '售卖规格',
    `service_year` int(1) DEFAULT NULL COMMENT '维保年限',
    `update_time`  datetime    DEFAULT NULL COMMENT '更新时间',
    `create_time`  datetime    DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品配置表';


CREATE TABLE `ip_demand_planning`
(
    `id`               BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `plan_id`          bigint(20) NOT NULL COMMENT '方案ID',
    `segment_type`     varchar(255) DEFAULT NULL COMMENT '网段类型',
    `vlan`             varchar(45)  DEFAULT NULL COMMENT 'VLAN ID',
    `c_num`            VARCHAR(45)  DEFAULT NULL COMMENT 'C数量',
    `address`          varchar(255) DEFAULT NULL COMMENT '地址段',
    `describe`         varchar(255) DEFAULT NULL COMMENT '描述',
    `address_planning` VARCHAR(255) DEFAULT NULL COMMENT 'IP地址规划建议',
    `create_time`      datetime     DEFAULT NULL COMMENT '创建时间',
    `update_time`      datetime     DEFAULT NULL COMMENT '更新时间',
    KEY                `IDX_PLAN_ID` (`plan_id`),
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPACT COMMENT='Ip需求规划表';


CREATE TABLE `network_model_role_rel`
(
    `network_device_role_id` bigint(20) DEFAULT NULL COMMENT '网络设备角色id',
    `network_model`          tinyint(4) DEFAULT NULL COMMENT '网络组网模式，1：三网合一，2：两网分离，3：三网分离',
    `associated_type`        tinyint(4) DEFAULT NULL COMMENT '关联类型，0：节点角色，1：网络设备角色',
    `role_id`                bigint(20) DEFAULT NULL COMMENT '关联的节点角色id或者网络设备角色id',
    `role_num`               int(11) DEFAULT NULL COMMENT '关联相同角色数量'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='网络组网模式与节点角色或者网络设备角色关联表';
