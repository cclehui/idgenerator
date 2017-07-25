CREATE TABLE `idGenerator` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',/*modifiable*/
  `worker_type` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '业务类型',/*modifiable*/
  `current_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '业务当前的递增id',/*modifiable*/
  PRIMARY KEY (`id`),
  KEY `idx_worker_type` (`worker_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='递增id保持表';
