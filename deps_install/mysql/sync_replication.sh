#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"

echo "=== Step 1: Stop slave on all slaves ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"

echo ""
echo "=== Step 2: Dump user_db from master and restore to slaves ==="
echo "Restoring user_db to mysql-1..."
sudo kubectl exec mysql-0 -n pay-ns -- mysqldump -u root -p${ROOT_PASSWORD} --databases user_db | sudo kubectl exec -i mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD}

echo "Restoring user_db to mysql-2..."
sudo kubectl exec mysql-0 -n pay-ns -- mysqldump -u root -p${ROOT_PASSWORD} --databases user_db | sudo kubectl exec -i mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD}

echo ""
echo "=== Step 3: Reset slave and start ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

echo ""
echo "=== Step 4: Verify slave status ==="
echo "--- mysql-1 ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"
echo "--- mysql-2 ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"

echo ""
echo "=== Step 5: Verify databases ==="
echo "--- mysql-0 (master) ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "--- mysql-1 (slave) ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "--- mysql-2 (slave) ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"

echo ""
echo "=== Step 6: Verify tables ==="
echo "--- mysql-0 user_db tables ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"
echo "--- mysql-1 user_db tables ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"
echo "--- mysql-2 user_db tables ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"