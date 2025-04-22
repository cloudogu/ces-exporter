package decrypt

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/decrypt/mocks"
	"github.com/cloudogu/cesapp-lib/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const pk = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA3HtyjR8l+0FyxX736c4o6bNq8T4b8M3DR1wza21tfkozOPCz
rD9m/ZaVLnHy1Y4cbRT67qNG/kbyZbcC0erJprzRi7LwXGJTUlhvP6hrkkozgsaQ
7BBxPTYO+aEneTWoZV4EiIafXKe4l63Og0N8PczNFB8QVQgjFu1Xe9oRCe8fX72L
loHCnK+i1Vm9i2npmDS60hRLYYZgxlmXk1KZ+l8B0hC0hLnouolMPRpl/o5MenDG
j978iaWiZKbYoyX/jzdcPe6ek6JmYloXoWzKSWWT7lbKW4nYp6O9hQ9+Y9hpq/jY
y3r6W4hN5X8VxnG1hDl+MA29VIFYFYOOODrgHQIDAQABAoIBAQDEEj//gci8FSrk
uRHs6Tp3AehDmxEr50AW8MaFbW3m1kORCnUt48BKGaSXBhyGj3d2BidVGvyiWiNs
EwE9/obPcbEDg+C+t24Tl7NvL+5hzPzb+ouccs7ROYa9tfOtlesoIiDz7IxB0KGW
uakiEFyndL6XezyB8deKpwyahoWKh2aXmT9xBn1bkMyila9gVPJ6FaZe1mVxCEBC
SkydhgwGrbNInJ2mQECNjPFpmDuX//LoZckLopo+trSZW5cRty7Q7YjoGlNQ5TVq
fz7Yjwa3IkA2y0R3UVza4xnoAmdClxyD91UOOB6UBnBirsK3PXvPatYgFxi0uiM4
8WyiXGhdAoGBAOzyhXOah67Rg4tdLQnX7K1WYbp6T0f69p/3+b15rf48cNwzUY4L
kGP8OsthmqKr/MVjy7lv6IST+1PcV3OqFWw6XhaOPgjM+/jmGEMo9BSYFt8hiZtq
YZuzTwuzWaM6MpSDfaBy2rnVsE8rM9Yj/3bm2ZkUYXttSMBWRD0vZ70TAoGBAO41
/0WgLHUvJkZtl3ZOkd9LOmCKhq2MDq/KpIm/rDNa5dvDVjUnVaIl0YO7bk+KDoM8
FV2o40B7FolxznRqZ1RmQTREZbMd8vCBq5LkKK5tO4FLD+isJI1Rl8TolvC/XWvv
Zru7i2oz5r5HvbB4Kt7VAd7v20n5GEpa44YP04QPAoGBANJUia/PyYeeRZWtVTB6
soY/uqqsrbmohcoEdnUCETgv8MMW7tsXWsnWeV5WOs0RvGR/rLTkKNOfBKcxXZO3
tCKJQUHmbByl0TnlDj53mQq64vqYq60A5roulgk94GDrZUC95ANMUOpLTKFKKU56
T+f9DcU7+Th2DvFk4lgpv31vAoGBAL73ZOUxalKrcNjHJMSAamsDSRJ6G0vn2yJM
pymTEn69IUbTyzmjhgAOp28fBGkZeVb2BP7n1P8tbjzTkro7TwkXTLCVIJ6+pLLw
kVaaOI7VHP4i6ecSkd8FCVGfUNpB36gW7VoVGMgUQahLpSNiwqOPSgeqbDdaTYHW
aU5hQ6U7AoGACov9j9BOvl6AiAOs1r/EXGtBSRscNeqePmjD/mTBTZfyIdac5y7v
C8E32bBvKtPJwCPiiujr8PfLW8wYWPsxbJaLhZSrFQzCyPJQC05atoh30fj9OVY7
VQ0a3RiA1TprwmXxb5Fx9cnr4Arz8BiOeBDCC/BJOOHP5piLxUJdl4E=
-----END RSA PRIVATE KEY-----
`

func Test_createKeyProvider(t *testing.T) {
	t.Run("create key provider", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   keyProviderKey,
				Value: "pkcs1v15",
			},
		}, nil)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("error receiving global config", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return(core.GlobalConfig{}, assert.AnError)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, provider)
	})

	t.Run("global config does not contain key provider key", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   "invalid",
				Value: "pkcs1v15",
			},
		}, nil)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.Error(t, err)
		assert.Nil(t, provider)
	})

	t.Run("invalid keyProvider", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   keyProviderKey,
				Value: "invalid",
			},
		}, nil)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

func TestGetKeyProvider(t *testing.T) {
	getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
	getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
		{
			Key:   keyProviderKey,
			Value: "pkcs1v15",
		},
	}, nil)

	provider, err := GetKeyProvider(getGlobalCfgMock.Execute)

	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// ensure clientOnce is only called once
	keyProviderOnce.Do(func() {
		keyProvider = nil
		errKeyProvider = assert.AnError
	})

	assert.NoError(t, err)
	assert.NotNil(t, keyProvider)
}

func TestCreateDecrypter(t *testing.T) {
	// fake KeyProviderOnceCall
	keyProviderOnce.Do(func() {})

	const doguName = "testDogu"

	t.Run("create decrypter", func(t *testing.T) {
		oldBaseVolumePath := doguVolumeBasePath
		tmpDir := os.TempDir()

		doguVolumeBasePath = tmpDir

		defer func() {
			doguVolumeBasePath = oldBaseVolumePath
		}()

		privateKeyPath := path.Join(tmpDir, fmt.Sprintf("/%s/volumes/_private", doguName))

		sErr := os.MkdirAll(privateKeyPath, 0755)
		require.NoError(t, sErr)

		defer func() {
			_ = os.RemoveAll(privateKeyPath)
		}()

		sErr = os.WriteFile(path.Join(privateKeyPath, "private.pem"), []byte(pk), 0666)
		require.NoError(t, sErr)

		keyProvider = &keys.KeyProvider{}
		errKeyProvider = nil

		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)

		decrypter, err := CreateDecrypter(doguName, getGlobalCfgMock.Execute)
		assert.NoError(t, err)
		assert.NotNil(t, decrypter)
	})

	t.Run("error getting key provider for decrypter", func(t *testing.T) {
		keyProvider = nil
		errKeyProvider = assert.AnError

		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)

		decrypter, err := CreateDecrypter(doguName, getGlobalCfgMock.Execute)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, Decrypter{}, decrypter)
	})

	t.Run("error creating private key", func(t *testing.T) {
		keyProvider = &keys.KeyProvider{}
		errKeyProvider = nil

		oldBaseVolumePath2 := doguVolumeBasePath

		defer func() {
			doguVolumeBasePath = oldBaseVolumePath2
		}()

		doguVolumeBasePath = "invalid"

		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)

		decrypter, err := CreateDecrypter(doguName, getGlobalCfgMock.Execute)
		assert.Error(t, err)
		assert.Equal(t, Decrypter{}, decrypter)
	})
}

func TestDecrypter_Decrypt(t *testing.T) {
	// fake KeyProviderOnceCall
	keyProviderOnce.Do(func() {})

	encString := "f3/qa3tEwRKuu1cKySUSNPSp3KjuhUm0QfWb+fauToskVf1qOvaJaHuLzXKDvUPWt4iCDa0FQ9+JI9E1UDecrTZnwX7/4lyzMVmS+yk4A97yKrD3DAHnOpskvThbbk+zPB3m29d6dhteJFPyJD254hoTVpx/UrzPN1pM7b/gQwvDo4cE920nHOuYSiCzB4bpQ6qKeAKEADUyXXcCIvhe6SxCZm2ZVuP+Il/bAR4lA0fFuWOwWq3+eikEfgckN9M5oVGmrYK16+QzTnbwLzwpchND36LH9LpCJvbZ2/Zu6pzbLW2K5raDEcNY+7wrDKT5j9DHgcrXEZQNDFze66yjdQ=="
	expDecryptedValue := "DecryptTest"

	provider, err := keys.NewKeyProvider("pkcs1v15")
	require.NoError(t, err)

	keyPair, err := provider.FromPrivateKey([]byte(pk))
	require.NoError(t, err)

	d := Decrypter{privateKey: keyPair.Private()}

	value, err := d.Decrypt(encString)
	assert.NoError(t, err)
	assert.Equal(t, expDecryptedValue, value)
}
