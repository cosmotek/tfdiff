install:
	go build -o $$GOBIN/tfdiff main.go

tape:
	vhs demo.tape

clean:
	-rm -r dist

macos:
	-mkdir -p dist/macos/arm64
	GOARCH=arm64 GOOS=darwin go build -o dist/macos/arm64/tfdiff main.go
	zip dist/tfdiff-macos-arm64 dist/macos/arm64/tfdiff

	-mkdir -p dist/macos/x86_64
	GOARCH=amd64 GOOS=darwin go build -o dist/macos/x86_64/tfdiff main.go
	zip dist/tfdiff-macos-x86_64 dist/macos/x86_64/tfdiff

linux:
	-mkdir -p dist/linux/arm64
	GOARCH=arm64 GOOS=linux go build -o dist/linux/arm64/tfdiff main.go
	zip dist/tfdiff-linux-arm64 dist/linux/arm64/tfdiff

	-mkdir -p dist/linux/x86_64
	GOARCH=amd64 GOOS=linux go build -o dist/linux/x86_64/tfdiff main.go
	zip dist/tfdiff-linux-x86_64 dist/linux/x86_64/tfdiff

windows:
	-mkdir -p dist/windows/x86_64
	GOARCH=amd64 GOOS=windows go build -o dist/windows/x86_64/tfdiff.exe main.go
	zip dist/tfdiff-windows-x86_64 dist/windows/x86_64/tfdiff.exe

release_binaries: clean macos linux windows
	