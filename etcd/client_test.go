package etcd

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.etcd.io/etcd/client/v2"
	"os"
	"slices"
	"testing"
)

func createMockMasterPath(t *testing.T, fileContent string) func() {
	oldMasterPath := nodeMasterPath
	mockMasterFile, err := os.CreateTemp(os.TempDir(), "client")
	if err != nil {
		t.Fatal(err)
	}

	_, err = mockMasterFile.WriteString(fileContent)
	if err != nil {
		t.Fatal(err)
	}

	nodeMasterPath = mockMasterFile.Name()

	reset := func() {
		nodeMasterPath = oldMasterPath
		_ = mockMasterFile.Close()
	}

	return reset
}

func Test_createEtcdClient(t *testing.T) {
	t.Run("create etcd client", func(t *testing.T) {
		reset := createMockMasterPath(t, `127.0.0.1
`)
		defer reset()

		eClient, err := createEtcdClient()
		assert.NoError(t, err)
		assert.NotNil(t, eClient)
	})

	t.Run("error reading nodemaster file", func(t *testing.T) {
		oldMasterPath := nodeMasterPath
		nodeMasterPath = "invalid"

		defer func() {
			nodeMasterPath = oldMasterPath
		}()

		eClient, err := createEtcdClient()
		assert.Error(t, err)
		assert.Nil(t, eClient)
	})

	t.Run("error parsing etcd url", func(t *testing.T) {
		reset := createMockMasterPath(t, `in valid
`)
		defer reset()

		eClient, err := createEtcdClient()
		assert.Error(t, err)
		assert.Nil(t, eClient)
	})
}

func Test_getEtcdClient(t *testing.T) {
	reset := createMockMasterPath(t, `127.0.0.1
`)
	defer reset()

	eClient, err := getEtcdClient()
	assert.NoError(t, err)
	assert.NotNil(t, eClient)

	// ensure clientOnce is only called once
	clientOnce.Do(func() {
		etcdClient = nil
		clientErr = assert.AnError
	})

	assert.NoError(t, err)
	assert.NotNil(t, eClient)
}

func createKeyValueStub() *client.Node {
	return &client.Node{
		Key: "config",
		Dir: true,
		Nodes: []*client.Node{
			{
				Key: "_global",
				Dir: true,
				Nodes: []*client.Node{
					{
						Key:   "logger",
						Dir:   false,
						Value: "defaultLogger",
					},
					{
						Key:   "environment",
						Dir:   false,
						Value: "test",
					},
				},
			},
		},
	}
}

func Test_getKeyValues(t *testing.T) {
	// fake ClientOnceCall
	clientOnce.Do(func() {})

	t.Run("get key values for config key", func(t *testing.T) {
		mockEtcdClient := mocks.NewKeysAPI(t)

		mockEtcdClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(
			&client.Response{
				Node: createKeyValueStub(),
			}, nil)

		etcdClient = mockEtcdClient
		clientErr = nil

		keyValues, err := getKeyValues("config")
		assert.NoError(t, err)
		assert.NotNil(t, keyValues)

		assert.Len(t, keyValues, 2)

		expValues := []string{"test", "defaultLogger"}

		for _, kv := range keyValues {
			assert.True(t, slices.Contains(expValues, kv.Value))
		}
	})

	t.Run("apply filter to keyValues received", func(t *testing.T) {
		mockEtcdClient := mocks.NewKeysAPI(t)

		mockEtcdClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(
			&client.Response{
				Node: createKeyValueStub(),
			}, nil)

		etcdClient = mockEtcdClient
		clientErr = nil

		keyValues, err := getKeyValues("config", func(kvs []core.KeyValue) []core.KeyValue {
			// test files that deletes every value
			assert.Len(t, kvs, 2)
			return []core.KeyValue{}
		})

		assert.NoError(t, err)
		assert.NotNil(t, keyValues)

		assert.Len(t, keyValues, 0)
	})

	t.Run("getEtcdClient returns error", func(t *testing.T) {
		clientErr = assert.AnError
		_, err := getKeyValues("config")

		assert.ErrorIs(t, err, clientErr)

		clientErr = nil
	})

	t.Run("error getting key value from etcd", func(t *testing.T) {
		mockEtcdClient := mocks.NewKeysAPI(t)
		mockEtcdClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		etcdClient = mockEtcdClient
		clientErr = nil

		_, err := getKeyValues("config")

		assert.ErrorIs(t, err, assert.AnError)
	})
}

func Test_isKeyNotFoundError(t *testing.T) {
	tests := []struct {
		inErr  error
		expect bool
	}{
		{
			inErr:  client.Error{Code: client.ErrorCodeKeyNotFound},
			expect: true,
		},
		{
			inErr:  client.Error{Code: client.ErrorCodeDirNotEmpty},
			expect: false,
		},
		{
			inErr:  assert.AnError,
			expect: false,
		},
		{
			inErr:  nil,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%+v", tt), func(t *testing.T) {
			assert.Equal(t, tt.expect, isKeyNotFoundError(tt.inErr))
		})
	}
}
