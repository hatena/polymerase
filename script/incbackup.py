#!/usr/bin/env python

import argparse
import subprocess
from string import Template
import json

try:
    from urllib.parse import urlparse, urlencode
    from urllib.request import urlopen, Request
    from urllib.error import HTTPError
except ImportError:
    from urlparse import urlparse
    from urllib import urlencode
    from urllib2 import urlopen, Request, HTTPError

cmd = Template("""
xtrabackup \
  --host ${mysql_host} \
  --port ${mysql_port} \
  --user ${mysql_user} \
  --password ${mysql_password} \
  --slave-info \
  --backup \
  --incremental-lsn=${last_lsn} \
  --stream=xbstream \
  | gzip -c \
  | curl -X POST -f \
    --data-binary \
    @- 'http://${xtralab_host}:${xtralab_port}/api/${db}/inc-backup/${last_lsn}'
""".strip())


def get_last_lsn(args):
    url = Template("http://${xtralab_host}:${xtralab_port}/api/${db}/last-lsn").substitute(vars(args))
    try:
        resp = urlopen(url)
        return json.loads(resp.read().decode('utf8'))['last_lsn']
    except HTTPError as e:
        print(e)
        return ""


def main():
    parser = argparse.ArgumentParser(description="full backup script")
    parser.add_argument('--mysql-host', dest='mysql_host', type=str, default='127.0.0.1')
    parser.add_argument('--mysql-port', dest='mysql_port', type=int, default=3306)
    parser.add_argument('--mysql-user', dest='mysql_user', type=str, required=True)
    parser.add_argument('--mysql-password', dest='mysql_password', type=str, required=True)
    parser.add_argument('--xtralab-host', dest='xtralab_host', type=str, required=True)
    parser.add_argument('--xtralab-port', dest='xtralab_port', type=int, default=80)
    parser.add_argument('--db', dest='db', type=str, required=True)
    args = parser.parse_args()

    last_lsn = get_last_lsn(args)
    if last_lsn == "":
        return
    d = vars(args)
    d['last_lsn'] = last_lsn
    subprocess.call(cmd.substitute(d), shell=True)

if __name__ == '__main__':
    main()
