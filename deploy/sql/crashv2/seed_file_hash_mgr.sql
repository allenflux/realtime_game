CREATE TABLE `seed_file_hash_mgr` (
                                      `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键ID',
                                      `channel_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '渠道ID,火箭：19999,爆点：19998',
                                      `term_start` bigint unsigned NOT NULL DEFAULT '0' COMMENT '起始期数ID',
                                      `term_end` bigint unsigned NOT NULL DEFAULT '0' COMMENT '结束期数ID',
                                      `file_name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件名',
                                      `file_hash` varchar(64) NOT NULL DEFAULT '' COMMENT '文件哈希值',
                                      `file_url` varchar(512) NOT NULL DEFAULT '' COMMENT '文件URL',
                                      `create_time` bigint NOT NULL DEFAULT '0' COMMENT '创建时间(UNIX时间戳)',
                                      PRIMARY KEY (`id`),
                                      KEY `idx_channel_term_range` (`channel_id`, `term_start`, `term_end`),
                                      KEY `idx_channel_file_hash` (`channel_id`, `file_hash`)
) ENGINE = InnoDB AUTO_INCREMENT = 101 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '种子文件哈希管理表'