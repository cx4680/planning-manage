ALTER TABLE `node_role_baseline` ADD COLUMN `support_multi_resource_pool` tinyint(1) DEFAULT NULL COMMENT '是否支持多资源池，0：否，1：是';

ALTER TABLE `server_planning` ADD COLUMN `mixed_resource_pool_id` bigint(20) DEFAULT NULL COMMENT '混合部署资源池id';