#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== Configure Semi-sync on Master ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL rpl_semi_sync_master_enabled=1;"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL rpl_semi_sync_master_timeout=10000;"

echo ""
echo "=== Configure Semi-sync on Slave 1 ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL rpl_semi_sync_slave_enabled=1;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL read_only=1;"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL super_read_only=1;"

echo ""
echo "=== Configure Semi-sync on Slave 2 ==="
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL rpl_semi_sync_slave_enabled=1;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL read_only=1;"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL super_read_only=1;"

echo ""
echo "=== Verify Semi-sync ==="
echo "Master:"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_master_enabled';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW STATUS LIKE 'Rpl_semi_sync_master_clients';"

echo ""
echo "Slave 1:"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';"
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'read_only';"

echo ""
echo "Slave 2:"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';"
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'read_only';"