# CES-Exporter


### Build & Run

```shell
docker build -t ces-exporter . 

docker run --name ces-exporter -d -p 7000:22 -v ./my-public-key.pub:/home/ces-exporter/.ssh/authorized_keys -v /my/data/to/sync:/data ces-exporter
```

### Sync data from exporter with rsync

```shell
rsync -avhz --delete -e "ssh -p 7000 -l ces-exporter -i /my-private-key" localhost:/data/ ./destination/
```