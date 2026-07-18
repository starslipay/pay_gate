#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== 1. 验证半同步复制 ==="
echo "--- Master (mysql-0) ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_master_enabled';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW STATUS LIKE 'Rpl_semi_sync_master_clients';"

echo ""
echo "--- Slave (mysql-1) ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';"

echo ""
echo "--- Slave (mysql-2) ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';"

echo ""
echo "=== 2. 验证只读设置 ==="
echo "--- Slave (mysql-1) read_only ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'read_only';"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'super_read_only';"

echo ""
echo "--- Slave (mysql-2) read_only ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'read_only';"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'super_read_only';"

echo ""
echo "=== 3. 测试数据同步 ==="
echo "--- 在 master 上创建测试表 ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "CREATE TABLE IF NOT EXISTS test_sync (id INT PRIMARY KEY, name VARCHAR(100));"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "INSERT INTO test_sync VALUES (1, 'test1'), (2, 'test2');"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"

echo ""
echo "--- 在 slave-1 上验证数据 ---"
sleep 2
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"

echo ""
echo "--- 在 slave-2 上验证数据 ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"

echo ""
echo "=== 4. 测试从库只读 ==="
echo "--- 在 slave-1 上尝试写入（应该失败）---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "INSERT INTO test_sync VALUES (3, 'should_fail');" 2>&1 || echo "✓ 从库写入被拒绝"

echo ""
echo "=== 验证完成 ==="