## Restore

`restore` サブコマンドはPolymeraseサーバにあるバックアップファイルからリストアするためのデータファイルを構築します。

### 例

```bash
$ polymerase restore --db test-db --from 2017-06-20 --apply-prepare
```

### オプション

- `--db` or `-d`: string MUST
    - リストアしたいバックアップのユニークキーを指定します
- `--from`: string MUST
    - リストアするバックアップの日付を指定します。
    - フェッチするバックアップは指定された日付以前です。
    - 形式は `YYYY-MM-dd` です。
- `--host`: string (default: 127.0.0.1)
    - Polymeraseサーバのホスト名です。
- `--port`: string (default: 24925)
    - Polymeraseサーバのポート番号です。
- `--max-bandwidth`: string
    - バックアップファイルのフェッチ時のバンド幅を制御します (bytes / sec)。
    - 形式はhuman-readableな数値を指定できます (例 10MB, 10KB, 100m)。
- `--use-innobackupex`: boolean
    - `xtrabackup` バイナリの代わりに `innobackupex` バイナリを使用します。
    - MySQLのバージョンが5.1以上5.5未満の場合にはこのオプションを有効にする必要があります。
    - (Future: MySQLのバージョンから自動的にこのオプションを有効にする予定)
