CREATE TABLE `task_position` (
  `id` int NOT NULL AUTO_INCREMENT,
  `task_name` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '任务名称',
  `position_id` bigint NOT NULL DEFAULT '0' COMMENT '当前任务跑到的偏移量id',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `task_name` (`task_name`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '不同任务跑批数据的偏移量id记录'