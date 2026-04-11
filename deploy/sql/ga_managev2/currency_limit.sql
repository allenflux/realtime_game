CREATE TABLE `currency_limit` (
                                  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                  `channel_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '游戏渠道ID',
                                  `client_id` varchar(64) NOT NULL DEFAULT '' COMMENT '渠道号',
                                  `currency` varchar(16) NOT NULL DEFAULT '' COMMENT '货币代码',
                                  `currency_precision` decimal(10, 2) NOT NULL DEFAULT '1.00' COMMENT '货币精度',
                                  `min_bet` int NOT NULL DEFAULT '10' COMMENT '最小投注',
                                  `max_bet` int NOT NULL DEFAULT '50000' COMMENT '最大投注',
                                  `max_profit` int NOT NULL DEFAULT '500000' COMMENT '最大盈利',
                                  `is_active` tinyint NOT NULL DEFAULT '2' COMMENT '是否启用 1=是 2=否',
                                  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                  PRIMARY KEY (`id`),
                                  UNIQUE KEY `uniq_channel_currency` (`channel_id`, `currency`),
                                  KEY `idx_client_id` (`client_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 2766 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '货币限额配置表'