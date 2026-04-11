CREATE TABLE `chain_progress` (
                                  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                  `channel_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '游戏火箭or爆点',
                                  `term_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '已发送到的期数ID（包含该期）',
                                  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                  `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                  PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 3 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '批量通道进度表'