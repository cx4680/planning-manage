ALTER TABLE `server_cap_planning` MODIFY COLUMN `number` DECIMAL(14,3) DEFAULT NULL COMMENT '数量';
ALTER TABLE `software_bom_planning` MODIFY COLUMN `number` DECIMAL(14,3) DEFAULT NULL COMMENT '数量';
ALTER TABLE `server_planning` ADD COLUMN `water_level` tinyint(4) DEFAULT NULL COMMENT '水位，百分数';