#!/bin/bash

# Launch MySQL server
/root/sandboxes/start_all

# Start benchmark

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
