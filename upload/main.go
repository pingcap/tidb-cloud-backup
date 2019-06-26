package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pingcap/tidb-cloud-backup/pkg"
)

var (
	cloud     string
	region    string
	bucket    string
	endpoint  string
	backupDir string
)

func init() {
	flag.StringVar(&cloud, "cloud", "", "Cloud storage to use")
	flag.StringVar(&region, "region", "", "The region to send requests to.")
	flag.StringVar(&bucket, "bucket", "tidb-backup", "Name of bucket")
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint of Ceph object store")
	flag.StringVar(&backupDir, "backup-dir", "", "Backup directory")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	b, err := pkg.SetupBucket(context.Background(), cloud, region, bucket, endpoint)
	if err != nil {
		log.Fatalf("Failed to setup bucket: %s", err)
	}

	prefixDir := fmt.Sprintf("%s/", filepath.Dir(backupDir))
	err = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		start := time.Now().Unix()
		log.Println("start to process file", "size=", humanize.Bytes(uint64(info.Size())), "path=", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}
		w, err := b.NewWriter(ctx, strings.TrimPrefix(path, prefixDir), nil)
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
		log.Println("upload file done", path)
		duration := time.Now().Unix() - start
		if duration-start > 120 {
			log.Println("slow upload file", path, duration)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("uploading failed: %v", err)
	}

}
