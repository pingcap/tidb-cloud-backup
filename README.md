# TiDB cloud backup

## Build

```shell
go build -o docker/uploader

docker build -t tennix/tidb-cloud-backup docker
```

## Run

Ref [here](https://cloud.google.com/docs/authentication/production#obtaining_and_providing_service_account_credentials_manually) to get google application credentials with cloud storage permission.

And then go to GCP console to create a cloud storage bucket.

```shell
ts=$(date +%Y-%m-%dT%H%M%S)

docker run -v $PWD/tidb_backup_${ts}:/backup tennix/tidb-cloud-backup mydumper \
    --outputdir=/backup \
    --host=<tidb-host> \
    --port=4000 \
    --user=root \
    --password=<password>

docker run -v $PWD/tidb_backup_${ts}:/tidb_backup_${ts} \
    -v /path/to/google-application-credentials:/gcp-credentials.json
    -e GOOGLE_APPLICATION_CREDENTIALS=/gcp-credentials.json
    tennix/tidb-cloud-backup uploader \
    --cloud=gcp \
    --bucket=<bucket-name> \
    --backup-dir=/tidb_backup_${ts}
```
