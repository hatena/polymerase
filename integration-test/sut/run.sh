#!/bin/bash

MASTER_PORT=23700
SLAVE_PORT=23701
DATABASE_NAME=testdb

# Launch MySQL server
/root/sandboxes/start_all

# Start benchmark
/root/sandboxes/rsandbox_mysql-5_6_35/m < /create_table.sql

ycsb load jdbc \
  -P /ycsb/workloads/workloada \
  -p db.driver=com.mysql.jdbc.Driver \
  -p db.url=jdbc:mysql://127.0.0.1:${MASTER_PORT}/${DATABASE_NAME} \
  -p db.user=msandbox \
  -p db.passwd=msandbox

# ycsb run jdbc -P workloads/workloada -p db.driver=com.mysql.jdbc.Driver -p db.url=jdbc:mysql://127.0.0.1:23700/testdb -p db.user=msandbox -p db.passwd=msandbox -target 10 &

sleep 5

# Do full-backup
polymerase \
  full-backup \
  --mysql-port ${SLAVE_PORT} \
  --mysql-user msandbox \
  --mysql-password msandbox \
  --db test-db \
  --host server

# Do inc-backup
polymerase \
  inc-backup \
  --mysql-port ${SLAVE_PORT} \
  --mysql-user msandbox \
  --mysql-password msandbox \
  --db test-db \
  --host server

# Do restore
