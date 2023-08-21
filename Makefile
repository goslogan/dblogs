all: macos linux windows 

windows:
	GOOARCH=amd64 GOOS=windows go build -ldflags="-s -w"; \
	tar -czvf windows-amd64-dbconfig.tar.gz dbconfig.exe README.md
	rm dbconfig.exe

macos: macos-intel macos-arm

macos-intel:
	GOOARCH=amd64 GOOS=darwin go build -ldflags="-s -w"; \
	chmod +x dbconfig; \
	tar -czvf macos-amd64-dbconfig.gz dbconfig README.md
	rm dbconfig

macos-arm:
	GOOARCH=arm64 GOOS=darwin go build -ldflags="-s -w"; \
	chmod +x dbconfig; \
	tar -czvf macos-arm64-dbconfig.gz dbconfig README.md
	rm dbconfig

linux:
	GOOARCH=amd64 GOOS=linux go build -ldflags="-s -w"; \
	chmod +x dbconfig; \
	tar -czvf linux-amd64-dbconfig.gz dbconfig README.md
	rm dbconfig
