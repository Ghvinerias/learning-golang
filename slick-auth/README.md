# Config Package

This package provides a simplified way to manage configurations by combining the functionality of `godotenv` and `viper`. It handles both environment variables and configuration files in a unified way.

## Features

- Load environment variables from `.env` files
- Read configuration from multiple sources (env vars, config files)
- Support for different config formats (JSON, YAML, TOML)
- Prefix-based environment variable filtering
- Type-safe configuration values
- Simple API with fluent initialization

## Installation

```bash
go get github.com/joho/godotenv
go get github.com/spf13/viper
```

## Usage

### Basic usage

```go
package main

import (
    "fmt"
    "log"
    
    "your-module/config"
)

func main() {
    // Initialize config with prefix
    cfg, err := config.New(
        config.WithEnvPrefix("APP"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Get values directly from environment or config file
    host := cfg.GetString("host")
    port := cfg.GetInt("port")
    debug := cfg.GetBool("debug")
    
    fmt.Println("Host:", host)
    fmt.Println("Port:", port)
    fmt.Println("Debug mode:", debug)
    
    // Get all environment variables with prefix
    allVars := cfg.GetAllWithPrefix("APP_")
    fmt.Println("All APP_ variables:", allVars)
}
```

### Using with a config file

```go
package main

import (
    "log"
    "your-module/config"
)

func main() {
    // Initialize config with config file
    cfg, err := config.New(
        config.WithEnvPrefix("APP"),
        config.WithConfigFile("config.yaml"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Rest of your code...
}
```

### Binding to a struct

```go
package main

import (
    "fmt"
    "log"
    "your-module/config"
)

type AppConfig struct {
    Server struct {
        Host string
        Port int
    }
    Database struct {
        URL      string
        Username string
        Password string
    }
    Debug bool
}

func main() {
    cfg, err := config.New(
        config.WithEnvPrefix("APP"),
        config.WithConfigFile("config.yaml"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    var appConfig AppConfig
    if err := cfg.Unmarshal(&appConfig); err != nil {
        log.Fatal("Failed to unmarshal config:", err)
    }
    
    fmt.Printf("Server: %s:%d\n", appConfig.Server.Host, appConfig.Server.Port)
    fmt.Printf("Database: %s\n", appConfig.Database.URL)
}
```

## Environment Variables

When using the `WithEnvPrefix` option, environment variables will be automatically bound to configuration keys. 
For example, with prefix "APP":

- `APP_SERVER_HOST` binds to `server.host`
- `APP_DATABASE_URL` binds to `database.url`

## Configuration File Formats

The package supports different configuration file formats through Viper:

- JSON
- YAML
- TOML
- HCL
- INI
- env file

The format is automatically detected from the file extension.
