CREATE TABLE `user_page_config` (
    `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
    `user_id` INT(10) NOT NULL DEFAULT '0' COMMENT '用户ID',
    `game_id` BIGINT(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT 'apisys游戏ID',
    `config_json` VARCHAR(16000) NOT NULL DEFAULT '' COMMENT '页面配置JSON',
    `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `uniq_user_game` (`user_id`, `game_id`) USING BTREE,
    INDEX `idx_user_id` (`user_id`) USING BTREE,
    INDEX `idx_game_id` (`game_id`) USING BTREE
) COMMENT='用户页面配置表'