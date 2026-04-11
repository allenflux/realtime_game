CREATE TABLE `currency_limit` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `channel_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '游戏渠道ID',
    `client_id` varchar(64) NOT NULL DEFAULT '' COMMENT '渠道号',
    `currency` varchar(16) NOT NULL DEFAULT '' COMMENT '货币代码',
    `currency_precision` decimal(10,2) NOT NULL DEFAULT '1.00' COMMENT '货币精度',
    `min_bet` int(10) NOT NULL DEFAULT 10 COMMENT '最小投注',
    `max_bet` int(10) NOT NULL DEFAULT 50000 COMMENT '最大投注',
    `max_profit` int(10) NOT NULL DEFAULT 500000 COMMENT '最大盈利',
    `is_active` TINYINT(3) NOT NULL DEFAULT '2' COMMENT '是否启用 1=是 2=否',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_channel_currency` (`channel_id`,`currency`),
    KEY `idx_client_id` (`client_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='货币限额配置表';

-- 插入channel_id=3的数据
INSERT INTO `currency_limit` (`channel_id`, `client_id`, `currency`, `currency_precision`, `min_bet`, `max_bet`, `max_profit`) VALUES
(3, '175', 'CNY', 0.01, 10, 50000, 500000),
(3, '175', 'VND', 0.1, 10, 50000, 500000),
(3, '175', 'THB', 0.01, 10, 50000, 500000),
(3, '175', 'TRY', 0.01, 10, 50000, 500000),
(3, '175', 'INR', 0.01, 10, 50000, 500000),
(3, '175', 'AUD', 0.01, 10, 50000, 500000),
(3, '175', 'PHP', 0.01, 10, 50000, 500000),
(3, '175', 'BRL', 0.01, 10, 50000, 500000),
(3, '175', 'MXN', 0.01, 10, 50000, 500000),
(3, '175', 'NPR', 0.01, 10, 50000, 500000),
(3, '175', 'BDT', 0.01, 10, 50000, 500000),
(3, '175', 'IDR', 0.01, 10, 50000, 500000),
(3, '175', 'KRW', 0.01, 10, 50000, 500000),
(3, '175', 'JPY', 0.01, 10, 50000, 500000),
(3, '175', 'PKR', 0.01, 10, 50000, 500000),
(3, '175', 'RON', 0.01, 10, 50000, 500000),
(3, '175', 'DKK', 0.01, 10, 50000, 500000),
(3, '175', 'NOK', 0.01, 10, 50000, 500000),
(3, '175', 'MMK', 0.1, 10, 50000, 500000),
(3, '175', 'TWD', 0.01, 10, 50000, 500000),
(3, '175', 'CUP', 0.01, 10, 50000, 500000),
(3, '175', 'USD', 0.01, 10, 50000, 500000),
(3, '175', 'EUR', 0.01, 10, 50000, 500000),
(3, '175', 'GBP', 0.01, 10, 50000, 500000),
(3, '175', 'CHF', 0.01, 10, 50000, 500000),
(3, '175', 'CAD', 0.01, 10, 50000, 500000),
(3, '175', 'RUB', 0.01, 10, 50000, 500000),
(3, '175', 'KES', 0.01, 10, 50000, 500000),
(3, '175', 'USDT', 0.01, 10, 50000, 500000),
(3, '175', 'KKC', 1.00, 10, 50000, 500000),
(3, '175', 'LUCK', 1.00, 10, 50000, 500000);

-- 插入channel_id=4的数据
INSERT INTO `currency_limit` (`channel_id`, `client_id`, `currency`, `currency_precision`, `min_bet`, `max_bet`, `max_profit`) VALUES
(4, '175', 'CNY', 0.01, 10, 50000, 500000),
(4, '175', 'VND', 0.1, 10, 50000, 500000),
(4, '175', 'THB', 0.01, 10, 50000, 500000),
(4, '175', 'TRY', 0.01, 10, 50000, 500000),
(4, '175', 'INR', 0.01, 10, 50000, 500000),
(4, '175', 'AUD', 0.01, 10, 50000, 500000),
(4, '175', 'PHP', 0.01, 10, 50000, 500000),
(4, '175', 'BRL', 0.01, 10, 50000, 500000),
(4, '175', 'MXN', 0.01, 10, 50000, 500000),
(4, '175', 'NPR', 0.01, 10, 50000, 500000),
(4, '175', 'BDT', 0.01, 10, 50000, 500000),
(4, '175', 'IDR', 0.01, 10, 50000, 500000),
(4, '175', 'KRW', 0.01, 10, 50000, 500000),
(4, '175', 'JPY', 0.01, 10, 50000, 500000),
(4, '175', 'PKR', 0.01, 10, 50000, 500000),
(4, '175', 'RON', 0.01, 10, 50000, 500000),
(4, '175', 'DKK', 0.01, 10, 50000, 500000),
(4, '175', 'NOK', 0.01, 10, 50000, 500000),
(4, '175', 'MMK', 0.1, 10, 50000, 500000),
(4, '175', 'TWD', 0.01, 10, 50000, 500000),
(4, '175', 'CUP', 0.01, 10, 50000, 500000),
(4, '175', 'USD', 0.01, 10, 50000, 500000),
(4, '175', 'EUR', 0.01, 10, 50000, 500000),
(4, '175', 'GBP', 0.01, 10, 50000, 500000),
(4, '175', 'CHF', 0.01, 10, 50000, 500000),
(4, '175', 'CAD', 0.01, 10, 50000, 500000),
(4, '175', 'RUB', 0.01, 10, 50000, 500000),
(4, '175', 'KES', 0.01, 10, 50000, 500000),
(4, '175', 'USDT', 0.01, 10, 50000, 500000),
(4, '175', 'KKC', 1.00, 10, 50000, 500000),
(4, '175', 'LUCK', 1.00, 10, 50000, 500000);

ALTER TABLE
    `currency_limit`
    ADD
        COLUMN `is_active` TINYINT(3) NOT NULL DEFAULT '2' COMMENT '是否启用 1=是 2=否'
AFTER
  `max_profit`,
  CHANGE COLUMN `min_bet` `min_bet` INT(10) NOT NULL DEFAULT 10 COMMENT '最小投注'
AFTER
  `currency_precision`,
  CHANGE COLUMN `max_bet` `max_bet` INT(10) NOT NULL DEFAULT 50000 COMMENT '最大投注'
AFTER
  `min_bet`,
  CHANGE COLUMN `max_profit` `max_profit` INT(10) NOT NULL DEFAULT 500000 COMMENT '最大盈利'
AFTER
  `max_bet`;