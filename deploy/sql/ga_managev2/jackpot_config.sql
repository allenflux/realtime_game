CREATE TABLE `jackpot_config` (
                                  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                  `channel_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '游戏渠道ID',
                                  `client_id` varchar(64) NOT NULL DEFAULT '' COMMENT '代理商ID',
                                  `game_name` varchar(30) NOT NULL DEFAULT '' COMMENT '游戏名称',
                                  `is_jackpot_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否开启奖池 0=否 1=是',
                                  `order_amount_ratio` decimal(5, 3) NOT NULL DEFAULT '1.000' COMMENT '奖池抽取订单金额比例（%）',
                                  `trigger_order_count` int NOT NULL DEFAULT '50' COMMENT '触发奖池需要的赛前投注单数',
                                  `prize_ratio_1st` decimal(5, 2) NOT NULL DEFAULT '20.00' COMMENT '第1名奖金比例（%）',
                                  `prize_ratio_2nd` decimal(5, 2) NOT NULL DEFAULT '10.00' COMMENT '第2名奖金比例（%）',
                                  `prize_ratio_3rd` decimal(5, 2) NOT NULL DEFAULT '5.00' COMMENT '第3名奖金比例（%）',
                                  `max_prize_multiple` int NOT NULL DEFAULT '10' COMMENT '奖金的最大投注倍数',
                                  `eat_all_divisor` int NOT NULL DEFAULT '100' COMMENT '通吃约数（用于计算最大抽取比例）',
                                  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                  PRIMARY KEY (`id`),
                                  UNIQUE KEY `uniq_channel_game` (`channel_id`, `game_name`)
) ENGINE = InnoDB AUTO_INCREMENT = 26 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '奖池配置表'