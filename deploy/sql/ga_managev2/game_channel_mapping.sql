CREATE TABLE `game_channel_mapping` (
                                        `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                        `client_id` int NOT NULL DEFAULT '0' COMMENT '代理商ID',
                                        `channel_id` varchar(32) NOT NULL DEFAULT '' COMMENT '游戏ID',
                                        `apisys_game_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'apisys游戏ID',
                                        `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                        `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                        `game_key` varchar(50) NOT NULL DEFAULT '' COMMENT 'apisys游戏key',
                                        PRIMARY KEY (`id`),
                                        UNIQUE KEY `uniq_client_game` (`client_id`, `apisys_game_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 86 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '游戏渠道映射表'