package main

import (
	"os"
	"strings"
	"fmt"
	"text/template"
)

type Application struct {
	ApplicationType      string
	ApplicationDNS       string
	ApplicationIP        string
	ApplicationPort      string
	ApplicationLogFolder string
	ApplicationSSLCer    string
	ApplicationSSLKey    string
}

func main() {
	// Define a template with custom functions
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"or": func(a, b bool) bool {
			return a || b
		},
	}

	tmpl := template.Must(template.New("template.txt").Funcs(funcMap).ParseFiles("template.txt"))

	applications := []Application{
		{
			ApplicationType:      "API",
			ApplicationDNS:       "Example.Test.API.Slick.ge",
			ApplicationIP:        "localhost",
			ApplicationPort:      "7015",
			ApplicationLogFolder: "Example.API",
			ApplicationSSLCer:    "/etc/nginx/ssl/01-06.API.Slick.ge.cer",
			ApplicationSSLKey:    "/etc/nginx/ssl/01-06.API.Slick.ge.key",
		},
		{
			ApplicationType:      "Web",
			ApplicationDNS:       "Example.Test.Web.Slick.ge",
			ApplicationIP:        "localhost",
			ApplicationPort:      "7020",
			ApplicationLogFolder: "Example.Web",
			ApplicationSSLCer:    "/etc/nginx/ssl/01-06.Web.Slick.ge.cer",
			ApplicationSSLKey:    "/etc/nginx/ssl/01-06.Web.Slick.ge.key",
		},
		{
			ApplicationType:      "Service",
			ApplicationDNS:       "Example.Test.Service.Slick.ge",
			ApplicationIP:        "localhost",
			ApplicationPort:      "7030",
			ApplicationLogFolder: "Example.Service",
			ApplicationSSLCer:    "/etc/nginx/ssl/01-06.Service.Slick.ge.cer",
			ApplicationSSLKey:    "/etc/nginx/ssl/01-06.Service.Slick.ge.key",
		},
	}

	// Execute the template with the data
	err := tmpl.Execute(os.Stdout, applications)
	fmt.Println("\\===================================================\\")
	if err != nil {
		panic(err)
	}

}
