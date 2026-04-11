1) 增加字段
ALTER TABLE `crash_term`
    ADD COLUMN `term_id` BIGINT UNSIGNED NOT NULL DEFAULT 0
  COMMENT '期数ID'
  AFTER `id`;

2) 历史数据回填为“每渠道从1开始连续”
MySQL 8.0（窗口函数）
WITH ranked AS (
  SELECT id,
         ROW_NUMBER() OVER (PARTITION BY channel_id ORDER BY id) AS rn
  FROM crash_term
)
UPDATE crash_term t
    JOIN ranked r ON r.id = t.id
    SET t.term_id = r.rn;

3) 检查是否有冲突/空值
-- 是否仍有未回填的
SELECT COUNT(*) FROM crash_term WHERE term_id = 0;

-- 检查重复（应为 0 行）
SELECT channel_id, term_id, COUNT(*) c
FROM crash_term
GROUP BY channel_id, term_id
HAVING c > 1;

4) 建唯一索引（并移除之前的普通索引）
-- 建立唯一约束：同一渠道同一期号仅一条
ALTER TABLE `crash_term`
    ADD UNIQUE KEY `uniq_channel_term_no` (`channel_id`, `term_id`);


-- 常规
ALTER TABLE `crash_term`
    ADD COLUMN `sha512_seed` VARCHAR(512)
        CHARACTER SET ascii COLLATE ascii_bin
        NOT NULL DEFAULT ''
    COMMENT 'sha512种子'
  AFTER `term_hash`;



ALTER TABLE `crash_term`
DROP INDEX `term_hash`,
DROP INDEX `idx_term_hash`,
  ADD COLUMN `term_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '期数ID' AFTER `id`,
  ADD COLUMN `sha512_seed` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'sha512种子' AFTER `term_id`,
  ADD UNIQUE KEY `uniq_channel_term_id` (`channel_id`, `term_id`);