FROM pingcap/tidb-enterprise-tools:latest

ARG VERSION=v1.48.0
RUN apk update && apk add ca-certificates
ADD backup.sh backup.sh
ADD bin/etcdbackuper /usr/local/bin/etcdbackuper

RUN \
  apk add pigz --no-cache \
  && chmod 755 backup.sh \
  && ln -s /backup.sh /usr/local/bin/uploader \
  && ln -s /backup.sh /usr/local/bin/downloader \
  && wget -nv https://github.com/ncw/rclone/releases/download/${VERSION}/rclone-${VERSION}-linux-amd64.zip \
  && unzip rclone-${VERSION}-linux-amd64.zip \
  && mv rclone-${VERSION}-linux-amd64/rclone /usr/local/bin/rclone \
  && chmod 755 /usr/local/bin/rclone \
  && rm -rf rclone-${VERSION}-linux-amd64.zip \
  && rm -rf rclone-${VERSION}-linux-amd64

RUN apk --no-cache add \
  python \
  py-pip \
  jq \
  bash \
  git \
  groff \
  less \
  mailcap \
  bash \
  && pip install --no-cache-dir awscli \
  && apk del py-pip \
  && rm -rf /var/cache/apk/* /root/.cache/pip/*
