## [v0.4.4](https://github.com/taku-k/polymerase/compare/v0.4.3...v0.4.4) (2017-07-18)

* Add options for backup and restore [#47](https://github.com/taku-k/polymerase/pull/47) ([taku-k](https://github.com/taku-k))
* Add cron command to generate cron file [#46](https://github.com/taku-k/polymerase/pull/46) ([taku-k](https://github.com/taku-k))
* Purge state in etcd when removing backups [#45](https://github.com/taku-k/polymerase/pull/45) ([taku-k](https://github.com/taku-k))
* Add test for restore command [#44](https://github.com/taku-k/polymerase/pull/44) ([taku-k](https://github.com/taku-k))

## [v0.4.3](https://github.com/taku-k/polymerase/compare/v0.4.2...v0.4.3) (2017-07-13)

* Not store intermediate restored files [#43](https://github.com/taku-k/polymerase/pull/43) ([taku-k](https://github.com/taku-k))
* Omit example lines in Gopkg.toml [#42](https://github.com/taku-k/polymerase/pull/42) ([taku-k](https://github.com/taku-k))
* Use github.com/fortytw2/leaktest and remove original leaktest [#41](https://github.com/taku-k/polymerase/pull/41) ([taku-k](https://github.com/taku-k))
* Use dep for dependency management instead of glide [#40](https://github.com/taku-k/polymerase/pull/40) ([taku-k](https://github.com/taku-k))
* Use StoreSpec for store-dir flag to validate in initializing [#39](https://github.com/taku-k/polymerase/pull/39) ([taku-k](https://github.com/taku-k))

## [v0.4.2](https://github.com/taku-k/polymerase/compare/v0.4.1...v0.4.2) (2017-07-11)

* Fix segmentation fault in case db is not found [#38](https://github.com/taku-k/polymerase/pull/38) ([taku-k](https://github.com/taku-k))
* Fix bug to restore backup [#37](https://github.com/taku-k/polymerase/pull/37) ([taku-k](https://github.com/taku-k))

## [v0.4.1](https://github.com/taku-k/polymerase/compare/v0.4.0...v0.4.1) (2017-07-07)

* Enrich information of nodes [#36](https://github.com/taku-k/polymerase/pull/36) ([taku-k](https://github.com/taku-k))
* Add decompress cmd flag for restore [#35](https://github.com/taku-k/polymerase/pull/35) ([taku-k](https://github.com/taku-k))

## [v0.4.0](https://github.com/taku-k/polymerase/compare/v0.3.1...v0.4.0) (2017-07-01)

* Release v0.4.0 [#33](https://github.com/taku-k/polymerase/pull/33) ([taku-k](https://github.com/taku-k))
* Changed default time format to include time zone [#32](https://github.com/taku-k/polymerase/pull/32) ([taku-k](https://github.com/taku-k))
* `info backups` command outputs as json [#31](https://github.com/taku-k/polymerase/pull/31) ([taku-k](https://github.com/taku-k))
* `info nodes` commands output as json [#30](https://github.com/taku-k/polymerase/pull/30) ([taku-k](https://github.com/taku-k))
* Add CHANGELOG.md [#29](https://github.com/taku-k/polymerase/pull/29) ([taku-k](https://github.com/taku-k))
* Add version command [#28](https://github.com/taku-k/polymerase/pull/28) ([taku-k](https://github.com/taku-k))

## [v0.3.1](https://github.com/taku-k/polymerase/compare/v0.3.0...v0.3.1) (2017-06-29)

* Stop travis-ci email notification on success [#27](https://github.com/taku-k/polymerase/pull/27) ([taku-k](https://github.com/taku-k))
* Use tar instead of pure go [#26](https://github.com/taku-k/polymerase/pull/26) ([taku-k](https://github.com/taku-k))
* Support integration test with transaction [#19](https://github.com/taku-k/polymerase/pull/19) ([taku-k](https://github.com/taku-k))
* Add --safe-slave-backup flag to xtrabackup command [#22](https://github.com/taku-k/polymerase/pull/22) ([taku-k](https://github.com/taku-k))

## [v0.3.0](https://github.com/taku-k/polymerase/compare/v0.2.0...v0.3.0) (2017-06-27)

* Add latest option for restore [#18](https://github.com/taku-k/polymerase/pull/18) ([taku-k](https://github.com/taku-k))
* Add integration test [#14](https://github.com/taku-k/polymerase/pull/14) ([taku-k](https://github.com/taku-k))
* Don't use glide for test rule [#16](https://github.com/taku-k/polymerase/pull/16) ([taku-k](https://github.com/taku-k))
* Add insecure-auth option for backup [#15](https://github.com/taku-k/polymerase/pull/15) ([taku-k](https://github.com/taku-k))
* Fix etcd client leak [#13](https://github.com/taku-k/polymerase/pull/13) ([taku-k](https://github.com/taku-k))
* Add info command [#12](https://github.com/taku-k/polymerase/pull/12) ([taku-k](https://github.com/taku-k))
* Restore backup information when restarting server [#11](https://github.com/taku-k/polymerase/pull/11) ([taku-k](https://github.com/taku-k))
* Add advertise-host option and record it [#10](https://github.com/taku-k/polymerase/pull/10) ([taku-k](https://github.com/taku-k))

## [v0.2.0](https://github.com/taku-k/polymerase/compare/v0.1.0...v0.2.0) (2017-06-21)

* Add allocator [#9](https://github.com/taku-k/polymerase/pull/9) ([taku-k](https://github.com/taku-k))

## [v0.1.0](https://github.com/taku-k/polymerase/compare/...v0.1.0) (2017-06-20)

* Add nodes command [#8](https://github.com/taku-k/polymerase/pull/8) ([taku-k](https://github.com/taku-k))
* Enable to handle restore subcommand in case of innobackupex [#7](https://github.com/taku-k/polymerase/pull/7) ([taku-k](https://github.com/taku-k))
* Fix link bug [#6](https://github.com/taku-k/polymerase/pull/6) ([taku-k](https://github.com/taku-k))
* Add docs [#5](https://github.com/taku-k/polymerase/pull/5) ([taku-k](https://github.com/taku-k))
* Change subcommand 'server' to 'start' [#4](https://github.com/taku-k/polymerase/pull/4) ([taku-k](https://github.com/taku-k))
* Change logrus to default log package [#2](https://github.com/taku-k/polymerase/pull/2) ([taku-k](https://github.com/taku-k))
* Add --max-bandwidth option for restore [#1](https://github.com/taku-k/polymerase/pull/1) ([taku-k](https://github.com/taku-k))

