CREATE TABLE `game_channel_mapping` (
    `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `client_id` int(11) NOT NULL DEFAULT 0 COMMENT '代理商ID',
    `channel_id` varchar(32) NOT NULL DEFAULT '' COMMENT '游戏ID',
    `apisys_game_id` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT 'apisys游戏ID',
    `game_key` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'apisys游戏key',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_client_game` (`client_id`, `apisys_game_id`)
) DEFAULT CHARSET=utf8mb4 COMMENT='游戏渠道映射表';