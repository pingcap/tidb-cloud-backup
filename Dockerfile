FROM pingcap/tidb-enterprise-tools:latest

RUN apk update && apk add ca-certificates
ADD bin/uploader /usr/local/bin/uploader
ADD bin/downloader /usr/local/bin/downloader
ADD bin/etcdbackuper /usr/local/bin/etcdbackuper
