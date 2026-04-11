CREATE TABLE `game_term_id` (
                                `id` bigint NOT NULL AUTO_INCREMENT,
                                `channel_id` int NOT NULL DEFAULT '0' COMMENT '游戏id',
                                `term_id` int NOT NULL DEFAULT '0' COMMENT '局id @crash_term.id',
                                `game_term_id` int NOT NULL DEFAULT '0' COMMENT '游戏下的局id',
                                PRIMARY KEY (`id`),
                                UNIQUE KEY `term_id` (`term_id`),
                                KEY `channel_id` (`channel_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 7491831 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '各个游戏下的连续自增期数id'