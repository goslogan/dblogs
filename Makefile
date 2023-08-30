all: macos linux windows 

windows:
	GOOARCH=amd64 GOOS=windows go build -ldflags="-s -w"; \
	tar -czvf windows-amd64-dblogs.tar.gz dblogs.exe README.md
	rm dblogs.exe

macos: macos-intel macos-arm

macos-intel:
	GOOARCH=amd64 GOOS=darwin go build -ldflags="-s -w"; \
	chmod +x dblogs; \
	tar -czvf macos-amd64-dblogs.gz dblogs README.md
	rm dblogs

macos-arm:
	GOOARCH=arm64 GOOS=darwin go build -ldflags="-s -w"; \
	chmod +x dblogs; \
	tar -czvf macos-arm64-dblogs.gz dblogs README.md
	rm dblogs

linux:
	GOOARCH=amd64 GOOS=linux go build -ldflags="-s -w"; \
	chmod +x dblogs; \
	tar -czvf linux-amd64-dblogs.gz dblogs README.md
	rm dblogs
