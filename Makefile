
all:
	echo "make in LINUX platform"
	GOOS=linux go build -o bin/uploader upload/main.go
	GOOS=linux go build -o bin/downloader download/main.go

