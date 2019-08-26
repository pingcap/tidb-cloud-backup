#!/bin/sh

set -e -x

export RUN_MODE=`basename "$0"`

for i in "$@"
do
case $i in
--cloud=*)
  export CLOUD="${i#*=}"
  shift
  ;;
--region=*)
  export REGION="${i#*=}"
  shift
  ;;
--bucket=*)
  export BUCKET="${i#*=}"
  shift
  ;;
--endpoint=*)
  export ENDPOINT="${i#*=}"
  shift
  ;;
--backup-dir=*|--srcDir=*)
  export BACKUP_DIR="${i#*=}"
  shift
  ;;
--destDir=*)
  export DEST_DIR="${i#*=}"
  shift
  ;;
  *)
  common::log "unknown option [$i]"
  exit 1
  ;;
esac

done

if [ "$CLOUD" = "gcp" ] && [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
  echo "GCP credentials path is not set"
  exit 1
fi

if [ "$CLOUD" = "ceph" ] || [ "$CLOUD" = "aws" ]; then
  if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "S3 access key is not set"
    exit 1
  fi
  if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "S3 access secret is not set"
    exit 1
  fi
fi

cat <<EOF > /tmp/rclone.conf
[aws]
type = s3
env_auth = false
provider = ${S3_PROVIDER:-"AWS"}
access_key_id = ${AWS_ACCESS_KEY_ID}
secret_access_key = ${AWS_SECRET_ACCESS_KEY}
region = ${REGION}
endpoint = ${ENDPOINT}
[ceph]
type = s3
env_auth = false
provider = ${S3_PROVIDER:-"Ceph"}
access_key_id = ${AWS_ACCESS_KEY_ID}
secret_access_key = ${AWS_SECRET_ACCESS_KEY}
region = :default-placement
endpoint = ${ENDPOINT}
[gcp]
type = google cloud storage
service_account_file = ${GOOGLE_APPLICATION_CREDENTIALS}
EOF

if [ "$RUN_MODE" = "uploader" ]; then
  export BACKUP_BASE_DIR="$(dirname "$BACKUP_DIR")"
  tar cvzf ${BACKUP_BASE_DIR}/backup.tgz -C ${BACKUP_BASE_DIR} $(basename "${BACKUP_DIR}")
  rclone --config /tmp/rclone.conf copyto ${BACKUP_BASE_DIR}/backup.tgz ${CLOUD}:${BUCKET}/${BACKUP_DIR}/backup.tgz
elif [ "$RUN_MODE" = "downloader" ]; then
  rclone --config /tmp/rclone.conf copyto ${CLOUD}:${BUCKET}/${BACKUP_DIR} ${DEST_DIR}
  if [ -f "$DEST_DIR/backup.tgz" ]; then
    tar xzvf ${DEST_DIR}/backup.tgz -C ${DEST_DIR} && rm ${DEST_DIR}/backup.tgz
  fi
else
  echo "Unknown run mode $RUN_MODE"
  exit 1
fi

