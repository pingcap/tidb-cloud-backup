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
    echo "unknown option [$i], accept options: cloud, region, bucket, endpoint, backup-dir, srcDir, destDir." >&2
    exit 1
    ;;
  esac
done

if [ "${CLOUD}" != "gcp" ] && [ "${CLOUD}" != "aws" ] && [ "${CLOUD}" != "ceph" ]; then
  echo "Cloud ${CLOUD:-"<empty>"} is not supported" >&2
  exit 1
fi

if [ "${CLOUD}" = "gcp" ] && [ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
  echo "GCP credentials path is not set" >&2
  exit 1
fi

if [ "${CLOUD}" = "ceph" ]; then
  if [ -z "${AWS_ACCESS_KEY_ID}" ]; then
    echo "S3 access key is not set" >&2
    exit 1
  fi
  if [ -z "${AWS_SECRET_ACCESS_KEY}" ]; then
    echo "S3 access secret is not set" >&2
    exit 1
  fi
fi

if [ "${CLOUD}" = "aws" ]; then
    if [ ! -z "${AWS_ACCESS_KEY_ID}" ] && [ ! -z "${AWS_SECRET_ACCESS_KEY}" ]; then
        export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
        export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
    fi
fi

cat <<EOF > /tmp/rclone.conf
[aws]
type = s3
env_auth = true
provider = ${S3_PROVIDER:-"AWS"}
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

# In tidb-backup job downloader and uploader take different dir parameters
# During uploading, the backup_dir is /<BASE_DIR>/<BACKUP_NAME>/, and the data will be uploaded to /<BACKUP_NAME>
# During downloading, the backup_dir is /<BACKUP_NAME>/, and the data will be downloaded from /<BACKUP_NAME>
export BACKUP_NAME=$(basename ${BACKUP_DIR})
if [ "${RUN_MODE}" = "uploader" ]; then
  export BACKUP_BASE_DIR=$(dirname ${BACKUP_DIR})
  tar -cf - ${BACKUP_NAME} -C ${BACKUP_BASE_DIR} | pigz -p 16 > ${BACKUP_BASE_DIR}/${BACKUP_NAME}.tgz
  rclone --config /tmp/rclone.conf copyto ${BACKUP_BASE_DIR}/${BACKUP_NAME}.tgz ${CLOUD}:${BUCKET}/${BACKUP_NAME}/${BACKUP_NAME}.tgz
  rm ${BACKUP_BASE_DIR}/${BACKUP_NAME}.tgz
elif [ "$RUN_MODE" = "downloader" ]; then
  rclone --config /tmp/rclone.conf copyto ${CLOUD}:${BUCKET}/${BACKUP_DIR} ${DEST_DIR}
  if [ -f "$DEST_DIR/${BACKUP_NAME}.tgz" ]; then
    tar -xzvf ${DEST_DIR}/${BACKUP_NAME}.tgz -C ${DEST_DIR} && rm ${DEST_DIR}/${BACKUP_NAME}.tgz
  fi
else
  echo "Unknown run mode ${RUN_MODE}" >&2
  exit 1
fi
