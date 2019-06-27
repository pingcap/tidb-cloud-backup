
all:
	echo "make in LINUX platform"
	GO111MODULE=on go build -o bin/uploader upload/main.go
	GO111MODULE=on go build -o bin/downloader download/main.go
	GO111MODULE=on go build -o bin/etcdbackuper etcdbackup/main.go
