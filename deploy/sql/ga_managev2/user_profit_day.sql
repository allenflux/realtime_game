CREATE TABLE `user_profit_day` (
                                   `id` int NOT NULL AUTO_INCREMENT,
                                   `user_id` int NOT NULL DEFAULT '0' COMMENT '用户id',
                                   `channel_id` int NOT NULL DEFAULT '0' COMMENT '游戏渠道id',
                                   `bet_amt` decimal(14, 2) NOT NULL DEFAULT '0.00' COMMENT '投注总额',
                                   `pupm_amt` decimal(14, 2) NOT NULL DEFAULT '0.00' COMMENT '抽水总额',
                                   `cashout_amt` decimal(14, 2) NOT NULL DEFAULT '0.00' COMMENT '兑现总额',
                                   `profit_amt` decimal(14, 2) NOT NULL DEFAULT '0.00' COMMENT '盈亏总额',
                                   `rate` int NOT NULL DEFAULT '0' COMMENT '总杀率 = 兑现总额/投注总额，字段值=实际值*1000000',
                                   `record_date` int NOT NULL DEFAULT '0' COMMENT '日期， 格式：YYYYMMDD, 如：20240604',
                                   `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                   PRIMARY KEY (`id`) USING BTREE,
                                   UNIQUE KEY `user_id` (`user_id`, `channel_id`, `record_date`) USING BTREE,
                                   KEY `user_record` (`user_id`, `record_date`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 13694 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '用户每日盈亏数据统计表'