package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/tennix/tidb-cloud-backup/pkg"
)

var (
	bucket    string
	backupDir string
	cloud     string
)

func init() {
	flag.StringVar(&cloud, "cloud", "", "Cloud storage to use")
	flag.StringVar(&bucket, "bucket", "tidb-backup", "Name of bucket")
	flag.StringVar(&backupDir, "backup-dir", "", "Backup directory")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	b, err := pkg.SetupBucket(context.Background(), cloud, bucket)
	if err != nil {
		log.Fatalf("Failed to setup bucket: %s", err)
	}

	base := filepath.Base(backupDir)
	err = filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read file: %s", err)
		}
		w, err := b.NewWriter(ctx, filepath.Join(base, info.Name()), nil)
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
		return nil
	})
	if err != nil {
		log.Fatalf("uploading failed: %v", err)
	}

}
