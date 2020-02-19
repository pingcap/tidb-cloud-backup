FROM pingcap/tidb-enterprise-tools:latest

ARG RCLONE_VERSION=v1.51.0
ARG DM_VERSION=v1.0.3
ARG TOOLKIT_VERSION=v3.0.9

RUN apk update && apk add ca-certificates
ADD backup.sh backup.sh

RUN \
  apk add pigz --no-cache \
  && chmod 755 backup.sh \
  && ln -s /backup.sh /usr/local/bin/uploader \
  && ln -s /backup.sh /usr/local/bin/downloader \
  && wget -nv https://github.com/ncw/rclone/releases/download/${RCLONE_VERSION}/rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && unzip rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && mv rclone-${RCLONE_VERSION}-linux-amd64/rclone /usr/local/bin/rclone \
  && chmod 755 /usr/local/bin/rclone \
  && rm -rf rclone-${RCLONE_VERSION}-linux-amd64.zip \
  && rm -rf rclone-${RCLONE_VERSION}-linux-amd64

RUN \
  wget -nv https://download.pingcap.org/dm-${DM_VERSION}-linux-amd64.tar.gz \
  && tar -xzf dm-${DM_VERSION}-linux-amd64.tar.gz \
  && mv dm-${DM_VERSION}-linux-amd64/bin/mydumper /mydumper \
  && chmod 755 /mydumper \
  && rm -rf dm-${DM_VERSION}-linux-amd64.tar.gz \
  && rm -rf dm-${DM_VERSION}-linux-amd64

# lightning can now support all features of loader in tidb-enterprise-tools
# we should replace base image to alpine once the deprecated loader is not used
RUN \
  wget -nv https://download.pingcap.org/tidb-toolkit-${TOOLKIT_VERSION}-linux-amd64.tar.gz \
  && tar -xzf tidb-toolkit-${TOOLKIT_VERSION}-linux-amd64.tar.gz \
  && mv tidb-toolkit-${TOOLKIT_VERSION}-linux-amd64/bin/tidb-lightning /tidb-lightning \
  && mv tidb-toolkit-${TOOLKIT_VERSION}-linux-amd64/bin/tidb-lightning-ctl /tidb-lightning-ctl \
  && chmod 755 /tidb-lightning /tidb-lightning-ctl \
  && rm -rf tidb-toolkit-${TOOLKIT_VERSION}-linux-amd64.tar.gz \
  && rm -rf tidb-toolkit-${TOOLKIT_VERSION}-linux-amd64
