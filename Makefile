install:
	go build -o $$GOBIN/tfdiff main.go

tape:
	vhs demo.tape