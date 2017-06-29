#!/bin/bash
set -ex

MASTER_PORT=23700
SLAVE_PORT=23701
DATABASE_NAME=testdb

benchmark() {
  /root/sandboxes/rsandbox_5_6_35/m < /create_table.sql

  ycsb load jdbc \
    -P /ycsb/workloads/workloadf \
    -p db.driver=com.mysql.jdbc.Driver \
    -p db.url=jdbc:mysql://127.0.0.1:${MASTER_PORT}/${DATABASE_NAME} \
    -p db.user=msandbox \
    -p db.passwd=msandbox

  ycsb run jdbc -target 10 \
    -P /ycsb/workloads/workloadf \
    -p db.driver=com.mysql.jdbc.Driver \
    -p db.url=jdbc:mysql://127.0.0.1:${MASTER_PORT}/${DATABASE_NAME} \
    -p db.user=msandbox \
    -p db.passwd=msandbox &
}

test_backup() {
  # Do full-backup
  polymerase \
    full-backup \
    --mysql-port ${SLAVE_PORT} \
    --mysql-user msandbox \
    --mysql-password msandbox \
    --db test-db \
    --host server

  sleep 30

  # Do inc-backup
  polymerase \
    inc-backup \
    --mysql-port ${SLAVE_PORT} \
    --mysql-user msandbox \
    --mysql-password msandbox \
    --db test-db \
    --host server
}

test_restore() {
  cd /

  polymerase \
    restore \
    --host server \
    --db test-db \
    --latest \
    --apply-prepare

  cd /sandbox && \
    make_sandbox 5.6.35 -- --no_show && \
    cd /root/sandboxes/msb_5_6_35 && \
    ./stop && \
    rm -r data && \
    mv /polymerase-restore/base data

  echo "server-id=10010" >> /root/sandboxes/msb_5_6_35/my.sandbox.cnf

  cat << EOF >> /root/sandboxes/msb_5_6_35/data/xtrabackup_slave_info
,master_host='127.0.0.1',
master_port=${MASTER_PORT},
master_user='msandbox',
master_password='msandbox';
EOF

  /root/sandboxes/msb_5_6_35/start

  /root/sandboxes/msb_5_6_35/use < /root/sandboxes/msb_5_6_35/data/xtrabackup_slave_info

  /root/sandboxes/msb_5_6_35/use -e "start slave"

  IO_IS_RUNNING=$(/root/sandboxes/msb_5_6_35/use -e "show slave status\G" | grep "Slave_IO_Running:" | awk '{ print $2 }')
  SQL_IS_RUNNING=$(/root/sandboxes/msb_5_6_35/use -e "show slave status\G" | grep "Slave_SQL_Running:" | awk '{ print $2 }')

  ## Check if IO thread is running ##
  if [ "$IO_IS_RUNNING" != "Yes" ]
  then
      echo -e "I/O thread for reading the master's binary log is not running (Slave_IO_Running)"
      exit 1
  fi

  ## Check for SQL thread ##
  if [ "$SQL_IS_RUNNING" != "Yes" ]
  then
      echo -e "SQL thread for executing events in the relay log is not running (Slave_SQL_Running)"
      exit 1
  fi
}

# Launch MySQL server
/root/sandboxes/start_all

# Start benchmark
benchmark

sleep 5

# Test backup commands
test_backup

# Do restore
test_restore
