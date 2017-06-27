#!/bin/bash

# Launch MySQL server
service mysql start

# Start benchmark

# Do full-backup
polymerase full-backup --mysql-user root --mysql-password root --db test-db --host server

# Do inc-backup

# Do restore
