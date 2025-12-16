package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const dataDir = "data"

func main() {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("File server listening on :9000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("read error:", err)
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		cmd := parts[0]

		switch cmd {
		case "PUT":
			handlePut(parts, reader, conn)
		case "GET":
			handleGet(parts, conn)
		case "DELETE":
			handleDelete(parts, conn)
		default:
			fmt.Fprintln(conn, "ERR unknown command")
		}
	}
}

func handlePut(parts []string, reader *bufio.Reader, conn net.Conn) {
	if len(parts) != 3 {
		fmt.Fprintln(conn, "ERR usage: PUT <filename> <size>")
		return
	}

	filename := filepath.Base(parts[1])
	size, err := strconv.Atoi(parts[2])
	if err != nil || size < 0 {
		fmt.Fprintln(conn, "ERR invalid size")
		return
	}

	path := filepath.Join(dataDir, filename)
	file, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(conn, "ERR cannot create file")
		return
	}
	defer file.Close()

	_, err = io.CopyN(file, reader, int64(size))
	if err != nil {
		fmt.Fprintln(conn, "ERR failed to read data")
		return
	}

	fmt.Fprintln(conn, "OK")
}

func handleGet(parts []string, conn net.Conn) {
	if len(parts) != 2 {
		fmt.Fprintln(conn, "ERR usage: GET <filename>")
		return
	}

	filename := filepath.Base(parts[1])
	path := filepath.Join(dataDir, filename)

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(conn, "ERR file not found")
		return
	}
	defer file.Close()

	info, _ := file.Stat()
	fmt.Fprintf(conn, "OK %d\n", info.Size())
	io.Copy(conn, file)
}

func handleDelete(parts []string, conn net.Conn) {
	if len(parts) != 2 {
		fmt.Fprintln(conn, "ERR usage: DELETE <filename>")
		return
	}

	filename := filepath.Base(parts[1])
	path := filepath.Join(dataDir, filename)

	if err := os.Remove(path); err != nil {
		fmt.Fprintln(conn, "ERR cannot delete file")
		return
	}

	fmt.Fprintln(conn, "OK")
}
