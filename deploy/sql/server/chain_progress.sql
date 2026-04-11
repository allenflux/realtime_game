CREATE TABLE `chain_progress` (
                                  `id`         int UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                                  `channel_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '游戏火箭or爆点',
                                  `term_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '已发送到的期数ID（包含该期）',
                                  `ctime`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                  `mtime`      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                  PRIMARY KEY (`id`)
) COMMENT='批量通道进度表';