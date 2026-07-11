#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== Revoke all permissions ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "REVOKE ALL PRIVILEGES ON *.* FROM 'starslipay'@'%';"

echo ""
echo "=== Grant only pay_db permissions ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "GRANT ALL PRIVILEGES ON pay_db.* TO 'starslipay'@'%';"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "FLUSH PRIVILEGES;"

echo ""
echo "=== Verify permissions ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW GRANTS FOR 'starslipay'@'%';"