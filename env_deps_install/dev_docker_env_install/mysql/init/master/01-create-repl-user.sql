-- 主库初始化: 安装半同步插件 + 创建复制用户
-- MySQL 8.0.26+ 使用新插件名 rpl_semi_sync_source (替代旧的 rpl_semi_sync_master)
-- 注意: INSTALL PLUGIN 只注册到 mysql.plugin 表, 插件在正式实例启动时才加载,
--        所以不能在 init SQL 中 SET GLOBAL (临时实例上插件尚未加载, 变量不存在)。
--        半同步变量通过 --init-file 在正式实例启动时设置。

-- 安装半同步主端插件 (使用新名 source, 对应 semisync_source.so)
INSTALL PLUGIN rpl_semi_sync_source SONAME 'semisync_source.so';

-- 创建 root@% 用户 (docker 镜像默认只创建 root@localhost, 容器间访问需要 root@%)
CREATE USER IF NOT EXISTS 'root'@'%' IDENTIFIED BY 'root123456';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;

-- 创建复制用户
CREATE USER IF NOT EXISTS 'repl'@'%' IDENTIFIED BY 'repl123456';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;
