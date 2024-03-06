<h1 align="center">
 <b>arangocado: a backup tool for ArangoDB non-enterprise users</b>
</h1>

A backups tool for [ArangoDB](https://arangodb.com/). Manage your backups in a centralized way.

---

## Features

- Dump / restore your database backups with support of remote S3 storage. Using [arangodump and arangorestore.](https://docs.arangodb.com/3.11/operations/backup-and-restore/#logical-backups)
- Automated creation of backups with cron like schedule
- Support of multiple storages per backup
- Support of storage schedule to use different approaches for uploading backups

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Config example](#config-example)

## Installation
```bash
docker pull mrlaker/arangocado
```
## Config example
```yaml
backups:
  - name: arangocado
    schedule: "0 * * * *"
    triggerImmediately: false
    host: "http+tcp://arango:8529"
    database: "arangocado"
    user: "root"
    password: ""
    historySize: 5
    collections:
      - some

    storages: [local, do]

  - name: "some_test"
    host: "http+tcp://arango:8529"
    database: "arangocado_test"
    user: "root"
    password: ""
    historySize: 5

    storages: [local]

storages:
  local:
    endpoint: "s3.somehost.com"
    accessKey: ""
    secretKey: ""
    bucket: "arangodb-backup"

  do:
    endpoint: "s3.someanotherhost.com"
    accessKey: ""
    secretKey: ""
    bucket: ""
    useSSL: true
    schedule: "0 */6 * * *"

```
