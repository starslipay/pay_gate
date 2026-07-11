#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== Setting semi-sync timeout to 0 (never degrade) ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SET GLOBAL rpl_semi_sync_master_timeout=0;"

echo ""
echo "=== Verify configuration ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_master_timeout';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW VARIABLES LIKE 'rpl_semi_sync_master_enabled';"