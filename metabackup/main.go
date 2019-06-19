package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pingcap/tidb-cloud-backup/pkg"
)

const (
	basicPath = "metadata_backup"
)

var (
	cloud        string
	region       string
	bucket       string
	cephEndpoint string
	etcdEndpoint string
)

func init() {
	flag.StringVar(&cloud, "cloud", "ceph", "Cloud storage to use")
	flag.StringVar(&region, "region", "", "The region to send requests to.")
	flag.StringVar(&bucket, "bucket", "tidb-meta-backup", "Name of bucket")
	flag.StringVar(&cephEndpoint, "cephEndpoint", "", "Endpoint of Ceph object store")
	flag.StringVar(&etcdEndpoint, "etcdEndpoint", "", "Endpoint of Etcd object store")
	flag.Parse()
}

func main() {
	now := time.Now().UTC()
	podName := os.Getenv("POD_NAME")
	etcdCtx := context.Background()

	tlsConfig, err := pkg.NewTLSConfig()
	if err != nil {
		log.Fatalf("Failed to create tls config: %v", err)
	}

	// etcd endpoint is divide by comma
	endpoints := strings.Split(etcdEndpoint, ",")

	// Save snapshot and return the byte array data
	bm := pkg.NewBackupManagerFromWriter(tlsConfig, endpoints)
	rev, data, err := bm.SaveSnap(etcdCtx)
	if err != nil {
		log.Fatalf("Failed to save snapshot: %v", err)
	}

	// Backup metadata from ETCD and uploader them to ceph directly
	cephCtx := context.Background()
	b, err := pkg.SetupBucket(cephCtx, cloud, region, bucket, cephEndpoint)
	if err != nil {
		log.Fatalf("Failed to setup bucket: %v", err)
	}

	// TODO POD_NAME is required, remove compatibility code later
	backupName, err := pkg.ResovleBackupFromPodName(podName)
	if err != nil {
		backupName = fmt.Sprintf(basicPath+"_v%d_%s", rev, now.Format("2006-01-02-15:04:05"))
	}

	w, err := b.NewWriter(cephCtx, backupName, nil)
	if err != nil {
		log.Fatalf("Failed to obtain writer: %s", err)
	}

	_, err = w.Write(data)
	if err != nil {
		log.Fatalf("Failed to write to bucket: %s", err)
	}

	if err = w.Close(); err != nil {
		log.Fatalf("Failed to close: %s", err)
	}
}
