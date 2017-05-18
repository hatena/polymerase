#!/usr/bin/env python

import argparse
import subprocess
from string import Template
import json
import copy
import tempfile
import os.path

try:
    from urllib.parse import urlparse, urlencode
    from urllib.request import urlopen, Request
    from urllib.error import HTTPError
except ImportError:
    from urlparse import urlparse
    from urllib import urlencode
    from urllib2 import urlopen, Request, HTTPError


def get_keys(args):
    url = Template("http://${xtralab_host}:${xtralab_port}/api/restore/search/${db}/${from}").substitute(vars(args))
    try:
        resp = urlopen(url)
        return json.loads(resp.read().decode('utf8'))['keys']
    except HTTPError as e:
        print(e)
        return None


def get_full_backup(args, storage, key, tempd):
    filename = os.path.join(tempd, 'base.tar.gz')
    get_backup(args, storage, key, filename)
    subprocess.call('mkdir {1}/base && tar xf {0} -C {1}/base'.format(filename, tempd), shell=True)


def get_inc_backup(args, storage, key, tempd, inc):
    filename = os.path.join(tempd, 'inc{0}.xb.gz'.format(inc))
    get_backup(args, storage, key, filename)
    subprocess.call(
        'gunzip -c {0} > {1}/inc{2}.xb && mkdir -p {1}/inc{2} && xbstream -x -C {1}/inc{2} < {1}/inc{2}.xb && rm -rf {1}/inc{2}.xb*'.format(
            filename, tempd, inc
        ), shell=True)


def get_backup(args, storage, key, filename):
    objs = copy.copy(vars(args))
    objs['storage'] = storage
    objs['key'] = key
    url = Template("http://${xtralab_host}:${xtralab_port}/api/restore/file/${storage}?key=${key}").substitute(objs)
    try:
        resp = urlopen(url)
        data = resp.read()
        with open(filename, 'wb') as f:
            f.write(data)
    except Exception as e:
        print(e)


def main():
    parser = argparse.ArgumentParser(description="full backup script")
    parser.add_argument('--data-dir', dest='data_dir', type=str)
    parser.add_argument('--xtralab-host', dest='xtralab_host', type=str, required=True)
    parser.add_argument('--xtralab-port', dest='xtralab_port', type=int, default=10109)
    parser.add_argument('--db', dest='db', type=str, required=True)
    parser.add_argument('--from', dest='from', type=str, required=True)
    args = parser.parse_args()

    keys = get_keys(args)
    if keys is None:
        return
    tempd = tempfile.mkdtemp()
    print(tempd)
    inc = len(keys) - 1
    idx = 0
    while inc > 0:
        key = keys[idx]
        get_inc_backup(args, key['storage_type'], key['key'], tempd, inc)
        idx += 1
        inc -= 1
    key = keys[len(keys) - 1]
    get_full_backup(args, key['storage_type'], key['key'], tempd)


if __name__ == '__main__':
    main()
