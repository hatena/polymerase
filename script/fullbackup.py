#!/usr/bin/env python

import argparse
import subprocess
from string import Template

cmd = Template("""
xtrabackup \
  --host ${mysql_host} \
  --port ${mysql_port} \
  --user ${mysql_user} \
  --password ${mysql_password} \
  --slave-info \
  --backup \
  --stream=tar \
  | gzip -c \
  | curl -X POST -f \
    --data-binary \
    @- 'http://${xtralab_host}:${xtralab_port}/api/full-backup/${db}'
""".strip())


def main():
    parser = argparse.ArgumentParser(description="full backup script")
    parser.add_argument('--mysql-host', dest='mysql_host', type=str, default='127.0.0.1')
    parser.add_argument('--mysql-port', dest='mysql_port', type=int, default=3306)
    parser.add_argument('--mysql-user', dest='mysql_user', type=str, required=True)
    parser.add_argument('--mysql-password', dest='mysql_password', type=str, required=True)
    parser.add_argument('--xtralab-host', dest='xtralab_host', type=str, required=True)
    parser.add_argument('--xtralab-port', dest='xtralab_port', type=int, default=10109)
    parser.add_argument('--db', dest='db', type=str, required=True)
    args = parser.parse_args()
    subprocess.call(cmd.substitute(vars(args)), shell=True)

if __name__ == '__main__':
    main()
