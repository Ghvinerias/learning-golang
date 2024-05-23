# ConfigGen

## Overview

`ConfigGen` is a Go application designed to generate NGINX configuration files based on specified command line flags. This project includes a `main.go` file and a `Makefile` to facilitate building the application for different operating systems.

## Prerequisites

- Go 1.16 or later
- Make (for using the Makefile)
- Git (for cloning the repository)

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/Ghvinerias/learning-golang.git
   cd learning-golang/ConfigGen
   ```

2. **Build the application:**

   You can build the application for different operating systems using the Makefile.

   For Windows:
   ```bash
   make build-win
   ```

   For Linux:
   ```bash
   make build-lin
   ```

   For both:
   ```bash
   make build-all
   ```

## Usage

### Command Line Flags

The application requires several command line flags to generate the NGINX configuration:

- `-Environment`: The environment (e.g., Test, Prod)
- `-Type`: The type of the application (e.g., API, Worker, WebSite)
- `-DNS_Node`: The DNS node
- `-DNS_NLB`: The DNS NLB
- `-App_Port`: The application port
- `-App_Name`: The application name
- `-SSL_Cert_CER`: The SSL certificate CER file
- `-SSL_Cert_KEY`: The SSL certificate KEY file
- `-App_Root`: The application root (for WebSite)

Example usage:
```bash
./ConfigGen -Environment=Prod -Type=API -DNS_Node=example.com -App_Port=8080 -App_Name=myApp -SSL_Cert_CER=myapp.cer -SSL_Cert_KEY=myapp.key
```

### Makefile Targets

The `Makefile` provides several targets to build and clean the project:

- `build-win`: Build the application for Windows
- `build-lin`: Build the application for Linux
- `build-all`: Build the application for both Windows and Linux
- `remove-win`: Remove the Windows build
- `remove-lin`: Remove the Linux build
- `remove-all`: Remove all builds

To build the application for Windows:
```bash
make build-win
```

To remove the Windows build:
```bash
make remove-win
```

To build the application for both Windows and Linux:
```bash
make build-all
```

To remove all builds:
```bash
make remove-all
```

## Example NGINX Configuration Generation

Based on the flags provided, the application will generate an NGINX configuration. For instance, running the following command:
```bash
./ConfigGen -Environment=Prod -Type=API -DNS_Node=example.com -App_Port=8080 -App_Name=myApp -SSL_Cert_CER=myapp.cer -SSL_Cert_KEY=myapp.key
```

Will produce an NGINX configuration similar to:
```nginx
server {
    listen 443 ssl;
    server_name example.com;
    location / {
        proxy_pass http://localhost:8080;
        proxy_redirect off;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
    }
    access_log /var/log/nginx/myApp/access.log main;
    error_log /var/log/nginx/myApp/error.log;
    ssl_certificate /etc/nginx/ssl/myapp.cer;
    ssl_certificate_key /etc/nginx/ssl/myapp.key;
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

