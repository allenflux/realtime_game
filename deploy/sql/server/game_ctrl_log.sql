CREATE TABLE `game_ctrl_log` (
  `id` int NOT NULL AUTO_INCREMENT,
  `client_id` int NOT NULL DEFAULT '0' COMMENT '渠道id',
  `term_id` int NOT NULL DEFAULT '0' COMMENT '游戏局数id',
  `user_id` int NOT NULL DEFAULT '0' COMMENT '此操作的管理员用户id',
  `is_control` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否控制 1=否 2=是',
  `break_payout_rate` int NOT NULL DEFAULT '0' COMMENT '爆点赔付比例 （*10000）',
  `user_profit_correct` int NOT NULL DEFAULT '0' COMMENT '用户盈亏修正比例（*10000）',
  `next_rand_multiple` int NOT NULL DEFAULT '0' COMMENT '下局随机爆点结果 (*100)',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建 时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `term_id` (`term_id`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE,
  KEY `client_id` (`client_id`) USING BTREE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = DYNAMIC COMMENT = '游戏配置控制日志记录表'