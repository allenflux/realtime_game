CREATE TABLE `game_cfg_log` (
                                `id` int NOT NULL AUTO_INCREMENT,
                                `client_id` int NOT NULL DEFAULT '0' COMMENT '渠道id',
                                `user_id` int NOT NULL DEFAULT '0' COMMENT '调整人uid',
                                `rake` int NOT NULL DEFAULT '0' COMMENT '抽水率（*10,000）',
                                `ctrl_trigger_rate` int NOT NULL DEFAULT '0' COMMENT '调控触发率（*10,000）',
                                `ctrl_put_rate` int NOT NULL DEFAULT '0' COMMENT '调控池输出率（*10,000）',
                                `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                PRIMARY KEY (`id`) USING BTREE,
                                KEY `client_id` (`client_id`) USING BTREE,
                                KEY `user_id` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC COMMENT = '游戏调控配置日志记录表'