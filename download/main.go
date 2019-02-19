package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/LinuxGit/tidb-cloud-backup/pkg"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
)

var (
	cloud    string
	bucket   string
	endpoint string
	srcDir   string
	destDir  string
)

func init() {
	flag.StringVar(&cloud, "cloud", "", "Cloud storage to use")
	flag.StringVar(&bucket, "bucket", "tidb-backup", "Name of bucket")
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint of Ceph object store")
	flag.StringVar(&srcDir, "srcDir", "", "Source data directory in bucket")
	flag.StringVar(&destDir, "destDir", "", "Destination directory on local")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	b, err := pkg.SetupBucket(context.Background(), cloud, bucket, endpoint)
	if err != nil {
		log.Fatalf("Failed to setup bucket: %s", err)
	}
	err = download(ctx, b, srcDir, destDir)
	if err != nil {
		log.Fatalf("Failed to download data from bucket: %s/%s to %s,error: %s", bucket, srcDir, destDir, err)
	}
}

func download(ctx context.Context, b *blob.Bucket, srcDir, destDir string) error {
	localBucket, err := fileblob.OpenBucket(destDir, nil)
	if err != nil {
		log.Fatal(err)
	}
	iter := b.List(&blob.ListOptions{Prefix: srcDir})
	for {
		obj, err := iter.Next(ctx)
		if err != nil {
			return err
		}
		if obj == nil {
			break
		}
		log.Println(fmt.Sprintf("Begin download file: %s", obj.Key))
		err = downloadFile(ctx, b, localBucket, obj.Key)
		if err != nil {
			log.Println(fmt.Sprintf("Download file: %s failed, error: %v", obj.Key, err))
			return err
		}
		log.Println(fmt.Sprintf("Download file: %s successfully", obj.Key))
	}
	return nil
}

func downloadFile(ctx context.Context, srcBucket *blob.Bucket, destBucket *blob.Bucket, file string) error {
	r, err := srcBucket.NewReader(ctx, file, nil)
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := destBucket.NewWriter(ctx, file, nil)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	return nil
}
