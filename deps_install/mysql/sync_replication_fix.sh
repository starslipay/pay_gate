#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"

echo "=== Step 1: Stop slave on all slaves ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"

echo ""
echo "=== Step 2: Dump user_db without GTID ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysqldump -u root -p${ROOT_PASSWORD} --databases user_db --set-gtid-purged=OFF --single-transaction > /tmp/user_db_dump.sql

echo ""
echo "=== Step 3: Copy and restore user_db to mysql-1 ==="
sudo kubectl cp mysql-0:/tmp/user_db_dump.sql /tmp/user_db_dump.sql -n pay-ns
sudo kubectl cp /tmp/user_db_dump.sql mysql-1:/tmp/user_db_dump.sql -n pay-ns
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} < /tmp/user_db_dump.sql

echo ""
echo "=== Step 4: Copy and restore user_db to mysql-2 ==="
sudo kubectl cp /tmp/user_db_dump.sql mysql-2:/tmp/user_db_dump.sql -n pay-ns
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} < /tmp/user_db_dump.sql

echo ""
echo "=== Step 5: Reset slave and start ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

echo ""
echo "=== Step 6: Verify slave status ==="
echo "--- mysql-1 ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"
echo "--- mysql-2 ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_IO_Running|Slave_SQL_Running"

echo ""
echo "=== Step 7: Verify databases ==="
echo "--- mysql-0 (master) ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "--- mysql-1 (slave) ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
echo "--- mysql-2 (slave) ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"

echo ""
echo "=== Step 8: Verify data ==="
echo "--- mysql-0 t_user_info ---"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SELECT * FROM t_uid_segment;"
echo "--- mysql-1 t_user_info ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SELECT * FROM t_uid_segment;"
echo "--- mysql-2 t_user_info ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SELECT * FROM t_uid_segment;"