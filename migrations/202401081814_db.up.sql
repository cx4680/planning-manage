CREATE TABLE `network_device_shelve`
(
    `id`                  bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `plan_id`             bigint(20) DEFAULT NULL COMMENT ' 方案id',
    `device_logical_id`   varchar(255) DEFAULT NULL COMMENT '网络设备逻辑ID',
    `device_id`           varchar(255) DEFAULT NULL COMMENT '网络设备ID',
    `sn`                  bigint(20) DEFAULT NULL COMMENT 'SN',
    `machine_room_abbr`   varchar(255) DEFAULT NULL COMMENT '机房缩写',
    `machine_room_number` varchar(255) DEFAULT NULL COMMENT '机房编号',
    `cabinet_number`      varchar(255) DEFAULT NULL COMMENT '机柜编号',
    `slot_position`       varchar(255) DEFAULT NULL COMMENT '槽位',
    `u_number`            int          DEFAULT NULL COMMENT 'U数',
    `create_user_id`      varchar(255) NULL DEFAULT NULL COMMENT '创建人id',
    `create_time`         datetime NULL DEFAULT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云产品基线';
