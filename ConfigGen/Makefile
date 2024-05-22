.SILENT :

build-win:
		GOOS=windows GOARCH=amd64 go build -o ./build/windows/ConfigGen.exe main.go
build-lin:
		GOOS=linux GOARCH=amd64 go build -o ./build/linux/ConfigGen main.go
build-all:
		GOOS=windows GOARCH=amd64 go build -o ./build/windows/ConfigGen.exe main.go
		GOOS=linux GOARCH=amd64 go build -o ./build/linux/ConfigGen main.go
remove-win:
		rm -rf ./build/windows/ConfigGen.exe
remove-lin:
		rm -rf ./build/linux/ConfigGen
remove-all:
		rm -rf ./build/windows/ConfigGen.exe
		rm -rf ./build/linux/ConfigGen
