CREATE TABLE `table_stats_monthly` (
                                       `stat_month` date NOT NULL,
                                       `db_name` varchar(128) NOT NULL,
                                       `table_name` varchar(128) NOT NULL,
                                       `row_count` bigint DEFAULT NULL,
                                       `size_gb` decimal(20, 4) DEFAULT NULL,
                                       PRIMARY KEY (`stat_month`, `db_name`, `table_name`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci