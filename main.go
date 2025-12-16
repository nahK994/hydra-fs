package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const tcpAddr = "localhost:9000" // Phase 1 TCP file server

func main() {
	r := gin.Default()

	r.POST("/files", uploadFile)
	r.GET("/files/:name", getFile)
	r.DELETE("/files/:name", deleteFile)

	log.Println("HTTP gateway listening on :8080")
	r.Run(":8080")
}

// POST /files (multipart/form-data)
func uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot connect to file server"})
		return
	}
	defer conn.Close()

	size := file.Size

	// Send PUT header
	io.WriteString(conn, "PUT "+file.Filename+" "+strconv.FormatInt(size, 10)+"\n")

	// Send file bytes
	io.Copy(conn, src)

	c.JSON(http.StatusOK, gin.H{"status": "uploaded"})
}

// GET /files/:name
func getFile(c *gin.Context) {
	name := c.Param("name")

	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot connect to file server"})
		return
	}
	defer conn.Close()

	io.WriteString(conn, "GET "+name+"\n")

	// Read header: OK <size>\n
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read error"})
		return
	}

	header := string(buf[:n])
	var size int64
	_, err = fmt.Sscanf(header, "OK %d", &size)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+name)
	c.Header("Content-Length", strconv.FormatInt(size, 10))
	c.Status(http.StatusOK)

	io.CopyN(c.Writer, conn, size)
}

// DELETE /files/:name
func deleteFile(c *gin.Context) {
	name := c.Param("name")

	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot connect to file server"})
		return
	}
	defer conn.Close()

	io.WriteString(conn, "DELETE "+name+"\n")

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
