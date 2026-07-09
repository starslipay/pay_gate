#!/bin/bash

ROOT_PASSWORD="root123456"

echo "=== Creating database ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "CREATE DATABASE IF NOT EXISTS pay_db;"

echo "=== Creating test table ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "CREATE TABLE IF NOT EXISTS test_sync (id INT PRIMARY KEY, name VARCHAR(100));"
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "INSERT INTO test_sync VALUES (1, 'test1'), (2, 'test2');"

echo "=== Verify master data ==="
sudo kubectl exec mysql-0 -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} pay_db -e "SELECT * FROM test_sync;"