CREATE TABLE `resource_pool` (
    `id`                 bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `plan_id`            bigint(20) DEFAULT NULL COMMENT '方案id',
    `node_role_id`       bigint(20) DEFAULT NULL COMMENT '节点角色id',
    `resource_pool_name` varchar(255) DEFAULT NULL COMMENT '资源池名称',
    `open_dpdk`          tinyint(1)  DEFAULT NULL COMMENT '是否开启DPDK，1：开启，0：关闭',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源池表';

ALTER TABLE `server_planning` ADD COLUMN `resource_pool_id` bigint(20) DEFAULT NULL COMMENT '资源池id';

ALTER TABLE `server_cap_planning` ADD COLUMN `resource_pool_id` bigint(20) DEFAULT NULL COMMENT '资源池id';