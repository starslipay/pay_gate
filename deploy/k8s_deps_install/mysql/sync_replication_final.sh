#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"

echo "=== Step 1: Stop slave ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"

echo ""
echo "=== Step 2: Disable read_only on slaves ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL super_read_only=0; SET GLOBAL read_only=0;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL super_read_only=0; SET GLOBAL read_only=0;"

echo ""
echo "=== Step 3: Sync user_db to slaves ==="
echo "--- Syncing to mysql-1 ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysqldump -u root -p${ROOT_PASSWORD} --databases user_db --set-gtid-purged=OFF | sudo kubectl exec -i mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD}

echo "--- Syncing to mysql-2 ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysqldump -u root -p${ROOT_PASSWORD} --databases user_db --set-gtid-purged=OFF | sudo kubectl exec -i mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD}

echo ""
echo "=== Step 4: Reset and restart replication ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

echo ""
echo "=== Step 5: Verify ==="
echo "--- Databases ---"
echo "mysql-0:"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "mysql-1:"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "mysql-2:"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"

echo ""
echo "--- Tables in user_db ---"
echo "mysql-0:"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"
echo "mysql-1:"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"
echo "mysql-2:"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"

echo ""
echo "--- Slave status ---"
echo "mysql-1:"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"
echo "mysql-2:"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"