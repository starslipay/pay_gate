-- --init-file: 每次正式实例启动时执行
-- 半同步插件已通过 docker-entrypoint-initdb.d 的 INSTALL PLUGIN 安装到 mysql.plugin 表,
-- 正式实例启动时会自动加载, 此处只需 SET GLOBAL 启用半同步变量。
-- MySQL 8.0.26+ 使用新变量名 rpl_semi_sync_source_enabled (替代旧的 rpl_semi_sync_master_enabled)

SET GLOBAL rpl_semi_sync_source_enabled = 1;
SET GLOBAL rpl_semi_sync_source_timeout = 0;  -- 永不降级为异步
