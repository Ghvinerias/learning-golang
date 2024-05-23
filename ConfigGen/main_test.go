package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestRunMain(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		shouldFail bool
	}{
		{
			name: "Test API Environment",
			args: []string{"--Environment", "Test", "--Type", "API", "--DNS_Node", "Example.Test.API.Slick.ge", "--App_Port", "7010", "--App_Name", "Example.API", "--SSL_Cert_CER", "API.Slick.ge.cer", "--SSL_Cert_KEY", "API.Slick.ge.key"},
		},
		{
			name: "Prod API Environment",
			args: []string{"--Environment", "Prod", "--Type", "API", "--DNS_Node", "Example.01.API.Slick.ge", "--DNS_NLB", "Example.NLB.API.Slick.ge", "--App_Port", "7010", "--App_Name", "Example.API", "--SSL_Cert_CER", "API.Slick.ge.cer", "--SSL_Cert_KEY", "API.Slick.ge.key"},
		},
		{
			name: "Test Worker Environment",
			args: []string{"--Environment", "Test", "--Type", "Worker", "--DNS_Node", "Example-worker.Test.Slick.ge", "--App_Port", "7022", "--App_Name", "SK.Example", "--SSL_Cert_CER", "test.Slick.ge.cer", "--SSL_Cert_KEY", "test.Slick.ge.key"},
		},
		{
			name: "Prod Worker Environment",
			args: []string{"--Environment", "Prod", "--Type", "Worker", "--DNS_Node", "Example-worker.01.Slick.ge", "--App_Port", "7022", "--App_Name", "SK.Example", "--SSL_Cert_CER", "Slick.ge.cer", "--SSL_Cert_KEY", "Slick.ge.key"},
		},
		{
			name: "Test WebSite Environment",
			args: []string{"--Environment", "Test", "--Type", "WebSite", "--DNS_Node", "Example.Test.Slick.ge", "--App_Root", "/opt/Services/Web-Sites/Example/Current", "--App_Name", "Example.Slick.ge", "--SSL_Cert_CER", "test.Slick.ge.cer", "--SSL_Cert_KEY", "test.Slick.ge.key"},
		},
		{
			name: "Prod WebSite Environment",
			args: []string{"--Environment", "Prod", "--Type", "WebSite", "--DNS_NLB", "Example.Slick.ge", "--DNS_Node", "Example.01.Slick.ge", "--App_Root", "/opt/Services/Web-Sites/Example/Current", "--App_Name", "Example.Slick.ge", "--SSL_Cert_CER", "Slick.ge.cer", "--SSL_Cert_KEY", "Slick.ge.key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil && !tt.shouldFail {
				t.Errorf("test %s failed: %v", tt.name, err)
			}
		})
	}
}
