#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"

echo "=== Step 1: Stop slave on all slaves ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"

echo ""
echo "=== Step 2: Get current master position ==="
MASTER_INFO=$(sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW MASTER STATUS" | tail -1)
echo "Master info: $MASTER_INFO"

echo ""
echo "=== Step 3: Dump data from master ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysqldump -u root -p${ROOT_PASSWORD} --all-databases --master-data=2 --single-transaction > /tmp/mysql_dump.sql

echo ""
echo "=== Step 4: Copy dump to slaves ==="
echo "Copying to mysql-1..."
sudo kubectl cp mysql-0:/tmp/mysql_dump.sql mysql-1:/tmp/mysql_dump.sql -n pay-ns
echo "Copying to mysql-2..."
sudo kubectl cp mysql-0:/tmp/mysql_dump.sql mysql-2:/tmp/mysql_dump.sql -n pay-ns

echo ""
echo "=== Step 5: Restore data on slaves ==="
echo "Restoring on mysql-1..."
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} < /tmp/mysql_dump.sql
echo "Restoring on mysql-2..."
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} < /tmp/mysql_dump.sql

echo ""
echo "=== Step 6: Start slave ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

echo ""
echo "=== Step 7: Verify slave status ==="
echo "--- mysql-1 ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"
echo "--- mysql-2 ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"

echo ""
echo "=== Step 8: Verify databases ==="
echo "--- mysql-0 (master) ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "--- mysql-1 (slave) ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "--- mysql-2 (slave) ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"