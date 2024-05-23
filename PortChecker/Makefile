.SILENT :
.PHONY: run-test build-all tidy-do

build-win:
		GOOS=windows GOARCH=amd64 go build -o ./build/windows/main.exe main.go
build-lin:
		GOOS=linux GOARCH=amd64 go build -o ./build/linux/main main.go
build-all:
		GOOS=windows GOARCH=amd64 go build -o ./build/windows/main.exe main.go
		GOOS=linux GOARCH=amd64 go build -o ./build/linux/main main.go
remove-win:
		rm -rf ./build/windows/main.exe
remove-lin:
		rm -rf ./build/linux/main
remove-all:
		rm -rf ./build/windows/main.exe
		rm -rf ./build/linux/main
run-test:
	go test -v
tidy-do:
		go mod tidy