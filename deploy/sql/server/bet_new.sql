ALTER TABLE `bet`
    ADD COLUMN `user_seed` VARCHAR(60)
        CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
        NOT NULL DEFAULT ''
    COMMENT '用户种子'
  AFTER `user_name`;

INSERT INTO `channel` (`id`, `client_id`, `game_name`, `is_active`, `channel_type`,`divisor`,`inc_num`)
VALUES (19998, '19998', 'chain_crash', 1, 19998,37,1.07161);
INSERT INTO `channel` (`id`, `client_id`, `game_name`, `is_active`, `channel_type`,`divisor`,`inc_num`)
VALUES (19999, '19999', 'chain_rocket', 1, 19999,37,1.07161);