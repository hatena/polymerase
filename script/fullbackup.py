#!/usr/bin/env python2

import argparse
import subprocess
from string import Template

cmd = Template("""
xtrabackup --host ${mysql_host} \
           --port ${mysql_port} \
           --user ${mysql_user} \
           --password ${mysql_passwd} \
           --slave-info \
           --backup \
           --stream=tar \
            | gzip -c \
             | curl -X POST \
                    --data-binary \
                    @- \
                    'http://${xtralab_host}:${xtralab_port}/api/v0/test-db/full-backup'
""".strip())


def main():
    parser = argparse.ArgumentParser(description="full backup script")
    parser.add_argument('--mysql-host', dest='mysql_host', type=str, default='127.0.0.1')
    parser.add_argument('--mysql-port', dest='mysql_port', type=int, default=3306)
    parser.add_argument('--mysql-user', dest='mysql_user', type=str, required=True)
    parser.add_argument('--mysql-passwd', dest='mysql_passwd', type=str, required=True)
    parser.add_argument('--xtralab-host', dest='xtralab_host', type=str, required=True)
    parser.add_argument('--xtralab-port', dest='xtralab_port', type=int, default=80)
    args = parser.parse_args()
    subprocess.call(cmd.substitute(vars(args)), shell=True)

if __name__ == '__main__':
    main()