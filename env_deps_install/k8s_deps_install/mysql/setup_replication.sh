#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"

echo "=== Step 1: Create replication user on master ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CREATE USER 'repl'@'%' IDENTIFIED BY '${REPL_PASSWORD}'; GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%'; FLUSH PRIVILEGES;"

echo "=== Step 2: Configure mysql-1 as slave ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1; START SLAVE;"

echo "=== Step 3: Configure mysql-2 as slave ==="
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1; START SLAVE;"

echo "=== Step 4: Check slave status ==="
echo "--- mysql-1 ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"
echo "--- mysql-2 ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"