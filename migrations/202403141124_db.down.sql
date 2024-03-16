DROP TABLE IF EXISTS `resource_pool`;

ALTER TABLE `server_planning` DROP COLUMN `resource_pool_id`;

ALTER TABLE `server_cap_planning` DROP COLUMN `resource_pool_id`;