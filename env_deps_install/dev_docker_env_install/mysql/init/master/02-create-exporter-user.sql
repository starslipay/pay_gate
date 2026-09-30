-- 创建 mysqld-exporter 监控用户
-- mysqld-exporter 需要 PROCESS（查看线程/锁）、REPLICATION CLIENT（查看主从状态）、SELECT（查询 performance_schema）
CREATE USER IF NOT EXISTS 'exporter'@'%' IDENTIFIED BY 'exporter123456';
GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO 'exporter'@'%';
-- performance_schema 权限（MySQL 8.0 需要显式授权）
GRANT SELECT ON performance_schema.* TO 'exporter'@'%';
FLUSH PRIVILEGES;
