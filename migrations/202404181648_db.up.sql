update `config_item` set `name` = '管理区域', `code` = 'manage' where `id` = 201;
update `config_item` set `name` = '容灾区域', `code` = 'disasterTolerance' where `id` = 203;
update `region_manage` set `type` = 'manage' where `type` = 'merge';