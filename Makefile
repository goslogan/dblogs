all: macos linux windows 

windows:
	GOOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -o windows-amd64-dbconfig; \
	gzip windows-amd64-dbconfig

macos: macos-intel macos-arm

macos-intel:
	GOOARCH=amd64 GOOS=darwin go build -ldflags="-s -w" -o macos-intel-dbconfig; \
	gzip macos-intel-dbconfig

macos-arm:
	GOOARCH=arm64 GOOS=darwin go build -ldflags="-s -w" -o macos-arm64-dbconfig; \
	gzip macos-arm64-dbconfig

linux:
	GOOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o linux-amd64-dbconfig; \
	gzip linux-amd64-dbconfig
