ALTER TABLE `channel`
    ADD COLUMN `channel_type` smallint UNSIGNED NOT NULL DEFAULT 0
  COMMENT '渠道类型：0：默认老玩法，一个渠道拥有独立的随机值开局结束，1：链上玩法，所有链上渠道共用同一个游戏期数、开始结束。19998：链上依赖的爆点渠道，19999：链上依赖的火箭渠道'