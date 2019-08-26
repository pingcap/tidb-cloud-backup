FROM pingcap/tidb-enterprise-tools:latest

ARG VERSION=v1.48.0
RUN apk update && apk add ca-certificates
ADD backup.sh backup.sh
# ADD bin/etcdbackuper /usr/local/bin/etcdbackuper

RUN \
  chmod 755 backup.sh \
  && cp backup.sh /usr/local/bin/uploader \
  && cp backup.sh /usr/local/bin/downloader \
  && wget -nv https://github.com/ncw/rclone/releases/download/${VERSION}/rclone-${VERSION}-linux-amd64.zip \
  && unzip rclone-${VERSION}-linux-amd64.zip \
  && mv rclone-${VERSION}-linux-amd64/rclone /usr/local/bin/rclone \
  && chmod 755 /rclone \
  && rm -rf rclone-${VERSION}-linux-amd64.zip
