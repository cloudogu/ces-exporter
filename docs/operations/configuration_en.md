# Configuration of the API key

## Configuration in Classic-CES
In order for the ces-exporter to be used correctly, an SSH public key and an API key must be stored.

The API key can be freely selected and enables access to the application's interfaces. The key entered here
must also be entered correctly in the importer for access to work.
To set the key for the exporter, this must simply be stored in etcd. The key responsible for this in
etcd is `/config/ces-exporter/authentication/api_key`.

The SSH key must be a valid public key and match the private key stored in the importer. It must have been generated
with a corresponding SSH command. In order to store this correctly in the exporter,
must be saved in etcd. The key responsible for this in etcd is
`/config/ces-exporter/authentication/public_key`.

## Using the api key

Requests to the exporter API need to be using the `X-CES-EXPORTER-API-KEY` header.