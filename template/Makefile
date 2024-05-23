.SILENT :
.PHONY: run-test build-all tidy-do

build-win:
		GOOS=windows GOARCH=amd64 go build -o ./build/windows/packageName.exe main.go
build-lin:
		GOOS=linux GOARCH=amd64 go build -o ./build/linux/packageName main.go
build-all:
		GOOS=windows GOARCH=amd64 go build -o ./build/windows/packageName.exe main.go
		GOOS=linux GOARCH=amd64 go build -o ./build/linux/packageName main.go
remove-win:
		rm -rf ./build/windows/*
remove-lin:
		rm -rf ./build/linux/*
remove-all:
		rm -rf ./build/windows/*
		rm -rf ./build/linux/*
run-test:
	go test -v
tidy-do:
		go mod tidy