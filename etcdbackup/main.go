package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/pingcap/tidb-cloud-backup/pkg"
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
	flag.StringVar(&region, "region", "", "The region to send requests to")
	flag.StringVar(&bucket, "bucket", "tidb-etcd-backup", "Name of bucket")
	flag.StringVar(&cephEndpoint, "cephEndpoint", "", "Endpoint of Ceph object store")
	flag.StringVar(&etcdEndpoint, "etcdEndpoint", "", "Endpoint of Etcd")
	flag.Parse()
}

func main() {
	etcdCtx := context.Background()
	pathSuffix := os.Getenv("PATH_SUFFIX")

	tlsConfig, err := pkg.NewTLSConfig()
	if err != nil {
		log.Fatalf("Failed to create tls config: %v", err)
	}

	// etcd endpoint is separated by comma
	endpoints := strings.Split(etcdEndpoint, ",")

	// Save snapshot and return the byte array data
	bm := pkg.NewBackupManagerFromWriter(tlsConfig, endpoints)
	rev, data, err := bm.SaveSnap(etcdCtx)
	if err != nil {
		log.Fatalf("Failed to save snapshot rev:%s err: %v", rev, err)
	}

	// Backup data from ETCD and uploader them to ceph directly
	cephCtx := context.Background()
	b, err := pkg.SetupBucket(cephCtx, cloud, region, bucket, cephEndpoint)
	if err != nil {
		log.Fatalf("Failed to setup bucket: %v", err)
	}

	backupName, err := pkg.ResovleBackupFromPathSuffix(pathSuffix)
	if err != nil {
		log.Fatalf("Failed to resolve path suffix : %v", err)
	}

	w, err := b.NewWriter(cephCtx, backupName, nil)
	if err != nil {
		log.Fatalf("Failed to obtain writer: %s", err)
	}
	defer w.Close()

	if _, err = w.Write(data); err != nil {
		log.Fatalf("Failed to write to bucket: %s", err)
	}
}
