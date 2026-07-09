#!/bin/bash

ROOT_PASSWORD="root123456"
REPL_PASSWORD="repl1reaucre123"
SLEEP_INTERVAL=10

echo "=== MySQL HA Monitor Started ==="

while true; do
  sleep $SLEEP_INTERVAL
  
  MASTER_POD=$(sudo kubectl get pods -n pay-ns -l mysql-role=master -o name 2>/dev/null | head -1 | cut -d'/' -f2)
  
  if [ -z "$MASTER_POD" ]; then
    echo "No master found, electing new master..."
    
    for POD in mysql-0 mysql-1 mysql-2; do
      if sudo kubectl exec $POD -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SELECT 1" 2>/dev/null; then
        echo "Elected $POD as new master"
        sudo kubectl label pods $POD -n pay-ns mysql-role=master --overwrite
        
        for OTHER_POD in mysql-0 mysql-1 mysql-2; do
          if [ "$OTHER_POD" != "$POD" ]; then
            echo "Setting $OTHER_POD as slave"
            sudo kubectl label pods $OTHER_POD -n pay-ns mysql-role=slave --overwrite
            
            if sudo kubectl exec $OTHER_POD -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SHOW SLAVE STATUS\G" 2>/dev/null | grep -q "Slave_IO_Running: No"; then
              echo "Reconfiguring replication for $OTHER_POD"
              sudo kubectl exec $OTHER_POD -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "STOP SLAVE; CHANGE MASTER TO MASTER_HOST='$POD.mysql.pay-ns.svc.cluster.local', MASTER_USER='repl', MASTER_PASSWORD='${REPL_PASSWORD}', MASTER_AUTO_POSITION=1; START SLAVE;"
            fi
          fi
        done
        break
      fi
    done
  else
    if ! sudo kubectl exec $MASTER_POD -n pay-ns -- mysql -u root -p${ROOT_PASSWORD} -e "SELECT 1" 2>/dev/null; then
      echo "Master $MASTER_POD is down, starting failover..."
      sudo kubectl label pods $MASTER_POD -n pay-ns mysql-role=slave --overwrite
    fi
  fi
  
  echo "Current state:"
  sudo kubectl get pods -n pay-ns -l app=mysql -o wide
done