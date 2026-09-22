-- 从库初始化: 安装半同步插件 + 配置复制源
-- MySQL 8.0.26+ 使用新插件名 rpl_semi_sync_replica (替代旧的 rpl_semi_sync_slave)
-- 注意: INSTALL PLUGIN 只注册到 mysql.plugin 表, 插件在正式实例启动时才加载,
--        所以不能在 init SQL 中 SET GLOBAL (临时实例上插件尚未加载, 变量不存在)。
--        半同步变量通过 --init-file 在正式实例启动时设置。
-- 正式实例启动时, 因默认 skip_slave_start=OFF, 会自动开始复制。

-- 安装半同步从端插件 (使用新名 replica, 对应 semisync_replica.so)
INSTALL PLUGIN rpl_semi_sync_replica SONAME 'semisync_replica.so';

-- 配置复制源 (使用 GTID 自动定位)
-- 此处只写入复制元数据, 不会立即连接主库 (临时实例阶段)
CHANGE REPLICATION SOURCE TO
  SOURCE_HOST='mysql-master',          -- docker-compose 服务名
  SOURCE_USER='repl',                  -- 主库创建的复制用户
  SOURCE_PASSWORD='repl123456',        -- 复制用户密码
  SOURCE_AUTO_POSITION=1;              -- 基于 GTID 自动定位复制位置
