#!/bin/bash

# Launch MySQL server
/root/sandboxes/start_all

# Start benchmark
# ycsb run jdbc -P workloads/workloada -p db.driver=com.mysql.jdbc.Driver -p db.url=jdbc:mysql://127.0.0.1:23700/testdb -p db.user=msandbox -p db.passwd=msandbox -target 10 &

# Do full-backup
polymerase \
  full-backup \
  --mysql-port 23701 \
  --mysql-user msandbox \
  --mysql-password msandbox \
  --db test-db \
  --host server

# Do inc-backup
polymerase \
  inc-backup \
  --mysql-port 23701 \
  --mysql-user msandbox \
  --mysql-password msandbox \
  --db test-db \
  --host server

# Do restore
