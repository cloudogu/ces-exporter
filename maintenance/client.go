package maintenance

import (
	"fmt"
	"go.etcd.io/etcd/client/v2"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// TODO reuse same types of configuration PR

const (
	_NodeMasterPath = "/etc/ces/node_master"
)

var (
	// singleton instance for etcd client
	etcdClient client.KeysAPI
	clientErr  error
	clientOnce sync.Once
)

func getEtcdEndpoint() (string, error) {
	nodeFile, err := os.ReadFile(_NodeMasterPath)
	if err != nil {
		return "", fmt.Errorf("failed to read node master file: %w", err)
	}

	slog.Debug("got node master file", "node_master", string(nodeFile))

	nMaster := strings.Trim(string(nodeFile), "\n")

	return fmt.Sprintf("http://%s:4001", nMaster), nil
}

func GetEtcdClient() (client.KeysAPI, error) {
	clientOnce.Do(func() {
		etcdEndpoint, err := getEtcdEndpoint()
		if err != nil {
			clientErr = fmt.Errorf("failed to get etcd endpoint: %w", err)
			return
		}

		cfg := client.Config{
			Endpoints: []string{etcdEndpoint},
			Transport: client.DefaultTransport,
		}

		c, err := client.New(cfg)
		if err != nil {
			clientErr = fmt.Errorf("failed to create etcd client: %w", err)
			return
		}

		etcdClient = client.NewKeysAPI(c)
	})

	return etcdClient, clientErr
}
