#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== Semi-sync Configuration ==="
echo "Master:"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_master_timeout';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW STATUS LIKE 'Rpl_semi_sync_master_clients';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW STATUS LIKE 'Rpl_semi_sync_master_status';"

echo ""
echo "Slave 1:"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';"

echo ""
echo "Slave 2:"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';"

echo ""
echo "=== Test Data Sync ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "INSERT INTO test_sync VALUES (3, 'test3');"
sleep 1
echo "Master:"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"
echo "Slave 1:"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"
echo "Slave 2:"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"