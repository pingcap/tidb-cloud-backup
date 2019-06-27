// Copyright 2017 The etcd-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

const (
	DefaultDialTimeout = 5 * time.Second
)

// BackupManager backups an etcd cluster.
type BackupManager struct {
	endpoints     []string
	etcdTLSConfig *tls.Config
}

// NewBackupManagerFromWriter creates a BackupManager with backup writer.
func NewBackupManagerFromWriter(tc *tls.Config, endpoints []string) *BackupManager {
	return &BackupManager{
		endpoints:     endpoints,
		etcdTLSConfig: tc,
	}
}

// SaveSnap uses backup writer to save etcd snapshot and return byte array
// and returns backup etcd server's kv store revision.
func (bm *BackupManager) SaveSnap(ctx context.Context) (int64, []byte, error) {
	etcdcli, rev, err := bm.etcdClientWithMaxRevision(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("create etcd client failed: %v", err)
	}
	defer etcdcli.Close()

	rc, err := etcdcli.Snapshot(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to receive snapshot (%v)", err)
	}
	defer rc.Close()

	// convert io.Reader to a byte array
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(rc)
	if err != nil {
		return 0, nil, fmt.Errorf("fail to covert reader to bytes (%v)", err)
	}

	return rev, buf.Bytes(), nil
}

// etcdClientWithMaxRevision gets the etcd endpoint with the maximum kv store revision
// and returns the etcd client of that member.
func (bm *BackupManager) etcdClientWithMaxRevision(ctx context.Context) (*clientv3.Client, int64, error) {
	etcdcli, rev, err := getClientWithMaxRev(ctx, bm.endpoints, bm.etcdTLSConfig)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get etcd client with maximum kv store revision: %v", err)
	}
	return etcdcli, rev, nil
}

func getClientWithMaxRev(ctx context.Context, endpoints []string, tc *tls.Config) (*clientv3.Client, int64, error) {
	mapEps := make(map[string]*clientv3.Client)
	var maxClient *clientv3.Client
	maxRev := int64(0)
	errors := make([]string, 0)
	for _, endpoint := range endpoints {
		cfg := clientv3.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: DefaultDialTimeout,
			TLS:         tc,
		}
		etcdcli, err := clientv3.New(cfg)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to create etcd client for endpoint (%v): %v", endpoint, err))
			continue
		}
		mapEps[endpoint] = etcdcli

		resp, err := etcdcli.Get(ctx, "/", clientv3.WithSerializable())
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to get revision from endpoint (%s)", endpoint))
			continue
		}

		if resp.Header.Revision > maxRev {
			maxRev = resp.Header.Revision
			maxClient = etcdcli
		}
	}

	// close all open clients that are not maxClient.
	for _, cli := range mapEps {
		if cli == maxClient {
			continue
		}
		cli.Close()
	}

	if maxClient == nil {
		return nil, 0, fmt.Errorf("could not create an etcd client for the max revision purpose from given endpoints (%v)", endpoints)
	}

	var err error
	if len(errors) > 0 {
		errorStr := ""
		for _, errStr := range errors {
			errorStr += errStr + "\n"
		}
		err = fmt.Errorf(errorStr)
	}

	return maxClient, maxRev, err
}
