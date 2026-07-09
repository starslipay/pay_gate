#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"

echo "=== Step 1: Stop slave on all slaves ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE;"

echo ""
echo "=== Step 2: Create database and tables on slaves ==="
echo "--- Creating on mysql-1 ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CREATE DATABASE IF NOT EXISTS user_db;"

echo "--- Creating on mysql-2 ---"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CREATE DATABASE IF NOT EXISTS user_db;"

echo ""
echo "=== Step 3: Get CREATE TABLE statements from master ==="
TABLES=$(sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;" | grep -v "Tables_in_user_db")

for TABLE in $TABLES; do
  echo "--- Creating table $TABLE on mysql-1 ---"
  CREATE_STMT=$(sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW CREATE TABLE $TABLE\G" | grep -A50 "Create Table:" | tail -n +2 | sed 's/^ //')
  sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; $CREATE_STMT;"
  
  echo "--- Creating table $TABLE on mysql-2 ---"
  sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; $CREATE_STMT;"
done

echo ""
echo "=== Step 4: Insert data from master to slaves ==="
for TABLE in $TABLES; do
  echo "--- Inserting data into $TABLE on mysql-1 ---"
  DATA=$(sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SELECT * FROM $TABLE;" | tail -n +2)
  # Simple INSERT for t_uid_segment
  if [ "$TABLE" = "t_uid_segment" ]; then
    while read -r LINE; do
      ID=$(echo $LINE | awk '{print $1}')
      UID_MAX=$(echo $LINE | awk '{print $2}')
      STEP=$(echo $LINE | awk '{print $3}')
      sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; INSERT INTO $TABLE (id, uid_max, step) VALUES ($ID, $UID_MAX, $STEP);"
    done <<< "$DATA"
    while read -r LINE; do
      ID=$(echo $LINE | awk '{print $1}')
      UID_MAX=$(echo $LINE | awk '{print $2}')
      STEP=$(echo $LINE | awk '{print $3}')
      sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; INSERT INTO $TABLE (id, uid_max, step) VALUES ($ID, $UID_MAX, $STEP);"
    done <<< "$DATA"
  fi
done

echo ""
echo "=== Step 5: Reset slave and start ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "RESET SLAVE;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CHANGE MASTER TO MASTER_HOST='mysql-0.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "START SLAVE;"

echo ""
echo "=== Step 6: Verify ==="
echo "--- Databases on slaves ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW DATABASES;"

echo ""
echo "--- Tables in user_db ---"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "USE user_db; SHOW TABLES;"