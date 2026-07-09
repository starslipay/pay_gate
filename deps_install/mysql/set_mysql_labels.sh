#!/bin/bash

echo "=== Setting MySQL role labels ==="

sudo kubectl label pods mysql-0 -n pay-ns mysql-role=master --overwrite
sudo kubectl label pods mysql-1 -n pay-ns mysql-role=slave --overwrite
sudo kubectl label pods mysql-2 -n pay-ns mysql-role=slave --overwrite

echo "=== Labels set ==="
sudo kubectl get pods -n pay-ns -l app=mysql --show-labels