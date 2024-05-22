package main

import (
    "flag"
    "fmt"
    "os"
    "text/template"
)

type Configuration struct {
    Environment string
    Type        string
    DNSNode     string
    DNSNLB      string
    AppPort     int
    AppName     string
    SSLCertCER  string
    SSLCertKEY  string
}

func main() {
    // Define command line flags
    environment := flag.String("Environment", "", "Environment (Test/Prod)")
    typeFlag := flag.String("Type", "", "Type (API/Worker/WebSite)")
    dnsNode := flag.String("DNS_Node", "", "DNS Node")
    dnsNLB := flag.String("DNS_NLB", "", "DNS NLB")
    appPort := flag.Int("App_Port", 0, "Application Port")
    appName := flag.String("App_Name", "", "Application Name")
    sslCertCER := flag.String("SSL_Cert_CER", "", "SSL Certificate CER")
    sslCertKEY := flag.String("SSL_Cert_KEY", "", "SSL Certificate KEY")

    // Parse command line flags
    flag.Parse()

    // Validate required flags
    if *environment == "" || *typeFlag == "" || *appPort == 0 || *appName == "" || *sslCertCER == "" || *sslCertKEY == "" {
        flag.PrintDefaults()
        os.Exit(1)
    }

    // Create configuration object
    config := Configuration{
        Environment: *environment,
        Type:        *typeFlag,
        DNSNode:     *dnsNode,
        DNSNLB:      *dnsNLB,
        AppPort:     *appPort,
        AppName:     *appName,
        SSLCertCER:  *sslCertCER,
        SSLCertKEY:  *sslCertKEY,
    }

    // Generate NGINX configuration
    nginxConfig := generateNginxConfig(config)

    fmt.Println(nginxConfig)
}


func nginx_templ_1(dnsNode, appPort, appName, sslCertCER, sslCertKEY string) string {
    return fmt.Sprintf(`
server {
    listen 443 ssl;
    server_name %s;
    location / {
        proxy_pass http://localhost:%s;
        proxy_redirect off;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
    }
    access_log /var/log/nginx/%s/access.log main;
    error_log /var/log/nginx/%s/error.log;
    ssl_certificate /etc/nginx/ssl/%s;
    ssl_certificate_key /etc/nginx/ssl/%s;
}
`, dnsNode, appPort, appName, appName, sslCertCER, sslCertKEY)
}

func nginx_templ_2(dnsNode, appRoot, appName, sslCertCER, sslCertKEY string) string {
    return fmt.Sprintf(`
server {
    listen 80;
    server_name %s;
    root %s;
    return 301 https://%s$request_uri;
}

server {
    listen 443 ssl;
    server_name %s;
    root %s;
    try_files $uri $uri/ /index.html;
    access_log /var/log/nginx/%s/access.log main;
    error_log /var/log/nginx/%s/error.log;
    ssl_certificate /etc/nginx/ssl/%s;
    ssl_certificate_key /etc/nginx/ssl/%s;
    proxy_buffer_size 128k;
    proxy_buffers 4 256k;
    proxy_busy_buffers_size 256k;
    large_client_header_buffers 4 32k;
    # client_max_body_size 10M;
    add_header Cache-Control 'private, no-store, no-cache, immutable, proxy-revalidate, max-age=0';
    expires 0;
}
`, dnsNode, appRoot, dnsNode, dnsNode, appRoot, appName, appName, sslCertCER, sslCertKEY)
}


serverBlock := nginx_templ_1("example.com", "8080", "myapp", "example.cer", "example.key")

