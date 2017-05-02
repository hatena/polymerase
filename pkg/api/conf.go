package api

// Most easiest way to configure
// an application is define config as
// yaml string and then parse it into
// map.
// How it works see here:
//     https://github.com/olebedev/config
var confString = `
debug: true
port: 5000
title: MySQL Backup API Server Integrated with Xtrabackup
api:
  prefix: /api
rootdir: /Users/taku_k/playground/mysql-backup-test/backups-dir
timeformat: 2006-01-02-15-04-05
`
