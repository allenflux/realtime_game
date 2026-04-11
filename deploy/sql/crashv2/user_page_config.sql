CREATE TABLE `user_page_config` (
                                    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',
                                    `user_id` int NOT NULL DEFAULT '0' COMMENT '用户ID',
                                    `game_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT 'apisys游戏ID',
                                    `config_json` varchar(16000) NOT NULL DEFAULT '' COMMENT '页面配置JSON',
                                    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                    PRIMARY KEY (`id`) USING BTREE,
                                    UNIQUE KEY `uniq_user_game` (`user_id`, `game_id`) USING BTREE,
                                    KEY `idx_user_id` (`user_id`) USING BTREE,
                                    KEY `idx_game_id` (`game_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 170501 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户页面配置表'