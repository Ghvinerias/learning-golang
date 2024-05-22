package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

func main() {
    if len(os.Args) < 4 {
        fmt.Println("Usage:")
        fmt.Println("  Listen mode: go run main.go listen <IP> <PORT>")
        fmt.Println("  Check mode:  go run main.go check <IP> <PORT>")
        return
    }

    mode := os.Args[1]
    ip := os.Args[2]
    port := os.Args[3]

    switch mode {
    case "listen":
        startListener(ip, port)
    case "check":
        checkConnection(ip, port)
    default:
        fmt.Println("Unknown mode:", mode)
        fmt.Println("Usage:")
        fmt.Println("  Listen mode: go run main.go listen <IP> <PORT>")
        fmt.Println("  Check mode:  go run main.go check <IP> <PORT>")
    }
}

func startListener(ip, port string) {
    address := net.JoinHostPort(ip, port)
    listener, err := net.Listen("tcp", address)
    if err != nil {
        fmt.Printf("Error starting TCP listener: %s\n", err)
        return
    }
    defer listener.Close()

    fmt.Printf("Listening on %s\n", address)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error accepting connection: %s\n", err)
            continue
        }

        fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err != net.ErrClosed {
                fmt.Printf("Error reading from connection: %s\n", err)
            }
            return
        }

        data := buffer[:n]
        fmt.Printf("Received: %s\n", string(data))

        // Echo the data back to the client
        _, err = conn.Write(data)
        if err != nil {
            fmt.Printf("Error writing to connection: %s\n", err)
            return
        }
    }
}

func checkConnection(ip, port string) {
    address := net.JoinHostPort(ip, port)

    timeout := 5 * time.Second
    conn, err := net.DialTimeout("tcp", address, timeout)
    if err != nil {
        fmt.Printf("Connection to %s failed: %s\n", address, err)
        return
    }
    defer conn.Close()

    fmt.Printf("Connection to %s succeeded\n", address)
}
