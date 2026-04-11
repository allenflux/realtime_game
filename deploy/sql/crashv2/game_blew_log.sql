CREATE TABLE `game_blew_log` (
                                 `id` int NOT NULL AUTO_INCREMENT,
                                 `client_id` int NOT NULL DEFAULT '0' COMMENT '渠道id',
                                 `term_id` int NOT NULL DEFAULT '0' COMMENT '游戏局数id',
                                 `user_id` int NOT NULL DEFAULT '0' COMMENT '引爆操作的管理员用户id',
                                 `manual_squib_state` tinyint NOT NULL DEFAULT '0' COMMENT '手动引爆阶段 0=未引爆 1=准备阶段 2=启动阶段 3=飞行阶段',
                                 `ctrl_result` tinyint(1) NOT NULL DEFAULT '1' COMMENT '引爆结果 1=成功 2=失败',
                                 `fail_msg` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '引爆失败原因',
                                 `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 PRIMARY KEY (`id`) USING BTREE,
                                 KEY `term_id` (`term_id`) USING BTREE,
                                 KEY `user_id` (`user_id`) USING BTREE,
                                 KEY `client_id` (`client_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC COMMENT = '游戏引爆操作日志记录表'