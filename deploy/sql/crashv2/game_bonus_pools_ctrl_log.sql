CREATE TABLE `game_bonus_pools_ctrl_log` (
                                             `id` int NOT NULL AUTO_INCREMENT,
                                             `client_id` int NOT NULL DEFAULT '0' COMMENT '渠道id',
                                             `user_id` int NOT NULL DEFAULT '0' COMMENT '此操作管理员的uid',
                                             `in_pools_amt` bigint NOT NULL DEFAULT '0' COMMENT '转入调控池金额 （*10000）',
                                             `out_pools_amt` bigint NOT NULL DEFAULT '0' COMMENT '调控池转出金额 （*10000）',
                                             `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                             PRIMARY KEY (`id`) USING BTREE,
                                             KEY `client_id` (`client_id`) USING BTREE,
                                             KEY `user_id` (`user_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC COMMENT = '奖池管理操作日志记录表'