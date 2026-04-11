CREATE TABLE `retry_cashout_task` (
                                      `id` int NOT NULL AUTO_INCREMENT,
                                      `bet_id` int NOT NULL DEFAULT '0' COMMENT '注单id@bet.id',
                                      `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '重试状态 1=待重试 2=重试中 3=已兑现 4=需要人工干预',
                                      `retry_num` tinyint(1) NOT NULL DEFAULT '0' COMMENT '已重试次数',
                                      `next_retry_time` bigint NOT NULL DEFAULT '0' COMMENT '下次重试时间戳',
                                      `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                      `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                      PRIMARY KEY (`id`) USING BTREE,
                                      UNIQUE KEY `bet_id` (`bet_id`) USING BTREE,
                                      KEY `status` (`status`, `next_retry_time`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 15214 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC COMMENT = '兑现失败需要重试的任务表'