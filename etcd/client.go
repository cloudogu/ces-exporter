package etcd

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"go.etcd.io/etcd/client/v2"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	nodeMasterPath = "/etc/ces/node_master"
)

var (
	// singleton instance for etcd client
	etcdClient client.KeysAPI
	clientErr  error
	clientOnce sync.Once
)

func getEtcdEndpoint() (string, error) {
	nodeFile, err := os.ReadFile(nodeMasterPath)
	if err != nil {
		return "", fmt.Errorf("failed to read node master file: %w", err)
	}

	slog.Debug("got node master file", "node_master", string(nodeFile))

	nMaster := strings.Trim(string(nodeFile), "\n")

	return fmt.Sprintf("http://%s:4001", nMaster), nil
}

func createEtcdClient() (client.KeysAPI, error) {
	etcdEndpoint, err := getEtcdEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to get etcd endpoint: %w", err)
	}

	cfg := client.Config{
		Endpoints: []string{etcdEndpoint},
		Transport: client.DefaultTransport,
	}

	c, err := client.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return client.NewKeysAPI(c), nil
}

func getEtcdClient() (client.KeysAPI, error) {
	clientOnce.Do(func() {
		etcdClient, clientErr = createEtcdClient()
	})

	return etcdClient, clientErr
}

type filterOption func(kvs []core.KeyValue) []core.KeyValue

func getKeyValues(path string, filters ...filterOption) ([]core.KeyValue, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("failed to get dir keys: %w", err)
	}

	c, err := getEtcdClient()
	if err != nil {
		return nil, wrapErr(err)
	}

	resp, err := c.Get(context.Background(), path, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, wrapErr(err)
	}

	keyValuePairs := extractValues(resp.Node, path)

	for _, filter := range filters {
		keyValuePairs = filter(keyValuePairs)
	}

	return keyValuePairs, nil
}

func extractValues(node *client.Node, excludePrefix string) []core.KeyValue {
	result := make([]core.KeyValue, 0)

	if !node.Dir {
		return []core.KeyValue{{Key: strings.TrimPrefix(node.Key, excludePrefix), Value: node.Value}}
	}

	for _, n := range node.Nodes {
		result = append(result, extractValues(n, excludePrefix)...)
	}

	return result
}

func isKeyNotFoundError(err error) bool {
	var etcdErr client.Error
	if errors.As(err, &etcdErr) {
		return etcdErr.Code == client.ErrorCodeKeyNotFound
	}

	return false
}
