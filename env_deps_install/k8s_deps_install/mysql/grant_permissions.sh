#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== Current user permissions ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW GRANTS FOR 'starslipay'@'%';"

echo ""
echo "=== Grant full permissions ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "GRANT ALL PRIVILEGES ON *.* TO 'starslipay'@'%' WITH GRANT OPTION;"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "FLUSH PRIVILEGES;"

echo ""
echo "=== Verify permissions ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW GRANTS FOR 'starslipay'@'%';"