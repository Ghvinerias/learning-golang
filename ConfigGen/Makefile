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
run-test-manual:
		go run main.go --Environment Test --Type API --DNS_Node Example.Test.API.Slick.ge --App_Port 7010 --App_Name Example.API --SSL_Cert_CER API.Slick.ge.cer --SSL_Cert_KEY API.Slick.ge.key
		go run main.go --Environment Prod --Type API --DNS_Node Example.01.API.Slick.ge --DNS_NLB Example.NLB.API.Slick.ge --App_Port 7010 --App_Name Example.API --SSL_Cert_CER API.Slick.ge.cer --SSL_Cert_KEY API.Slick.ge.key
		go run main.go --Environment Test --Type Worker --DNS_Node Example-worker.Test.Slick.ge --App_Port 7022 --App_Name SK.Example --SSL_Cert_CER test.Slick.ge.cer --SSL_Cert_KEY test.Slick.ge.key
		go run main.go --Environment Prod --Type Worker --DNS_Node Example-worker.01.Slick.ge --App_Port 7022 --App_Name SK.Example --SSL_Cert_CER Slick.ge.cer --SSL_Cert_KEY Slick.ge.key
		go run main.go --Environment Test --Type WebSite --DNS_Node Example.Test.Slick.ge --App_Root /opt/Services/Web-Sites/Example/Current --App_Name Example.Slick.ge --SSL_Cert_CER test.Slick.ge.cer --SSL_Cert_KEY test.Slick.ge.key
		go run main.go --Environment Prod --Type WebSite --DNS_NLB Example.Slick.ge --DNS_Node Example.01.Slick.ge --App_Root /opt/Services/Web-Sites/Example/Current --App_Name Example.Slick.ge --SSL_Cert_CER Slick.ge.cer --SSL_Cert_KEY Slick.ge.key