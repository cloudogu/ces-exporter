# CES-Exporter


### Build & Run

```shell
docker build -t ces-exporter . 
docker build -t registry.cloudogu.com/testing/ces-exporter:0.0.1 . 

docker push registry.cloudogu.com/testing/ces-exporter:0.0.1
docker pull registry.cloudogu.com/testing/ces-exporter:0.0.1

docker run --name ces-exporter -d -p 7000:22 -v ./my-public-key.pub:/root/.ssh/authorized_keys -v /my/data/to/sync:/data ces-exporter
docker run --name ces-exporter -d -p 7000:22 -v /home/bernst/ces_migrate_key.pub:/root/.ssh/authorized_keys -v /var/lib/ces:/data registry.cloudogu.com/testing/ces-exporter:0.0.1
```

### Sync data from exporter with rsync

```shell
rsync -avhz --delete -e "ssh -p 7000 -l ces-exporter -i /my-private-key" localhost:/data/ ./destination/
```