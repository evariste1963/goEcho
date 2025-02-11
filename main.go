package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket connection handler
func wsHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return err
	}
	defer conn.Close()

	clients[conn] = true
	log.Println("New WebSocket client connected")

	// Keep connection open
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}
	}
	return nil
}

// Notify clients about changes
func broadcastChange(fileType string) {
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(fileType))
		if err != nil {
			log.Println("WebSocket send error:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}

// Watch for file changes
func watchFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Directories to watch
	paths := []string{"views", "public"}

	for _, path := range paths {
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				return watcher.Add(p)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			ext := filepath.Ext(event.Name)
			switch ext {
			case ".css":
				log.Println("CSS changed, reloading styles...")
				broadcastChange("css")
			case ".js":
				log.Println("JS changed, reloading scripts...")
				broadcastChange("js")
			case ".html", ".go":
				log.Println("Go or HTML file changed, reloading page...")
				broadcastChange("reload")
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("File watcher error:", err)
		}
	}
}

func main() {
	e := echo.New()

	// Serve static files (CSS, JS, etc.)
	e.Static("/public", "public")

	// WebSocket for HMR
	e.GET("/ws", wsHandler)

	// Serve HTML page
	e.GET("/", func(c echo.Context) error {
		return c.File("views/index.html")
	})

	// Start file watcher in the background
	go watchFiles()

	e.Logger.Fatal(e.Start(":8080"))
}
