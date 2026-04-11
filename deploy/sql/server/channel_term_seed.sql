CREATE TABLE `channel_term_seed` (
                                     `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键ID',
                                     `channel_id` INT NOT NULL DEFAULT 0 COMMENT '渠道ID',
                                     `term_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '期数ID',
                                     `seed_id` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '种子ID',
                                     `seed_status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0-未使用 1-已使用',
                                     `ctime` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间(UNIX时间戳)',
                                     PRIMARY KEY (`id`),
                                     UNIQUE KEY `uniq_channel_term` (`channel_id`, `term_id`),
                                     KEY `idx_seed_id` (`seed_id`)
) DEFAULT CHARSET=utf8mb4 COMMENT='渠道-期数-种子 映射表'