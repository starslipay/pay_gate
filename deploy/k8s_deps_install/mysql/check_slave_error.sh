#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== mysql-1 ==="
sudo kubectl exec mysql-1 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_SQL_Running|Last_SQL_Error|Last_Error"

echo ""
echo "=== mysql-2 ==="
sudo kubectl exec mysql-2 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" | grep -E "Slave_SQL_Running|Last_SQL_Error|Last_Error"