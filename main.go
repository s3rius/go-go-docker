package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"gopkg.in/olahol/melody.v1"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//type ContainerBroadcaster struct {
//	subCount   int
//	containers chan types.ContainerJSON
//	stop        chan bool
//}

var containers = map[string]*ContainerBroadcaster{}

var infoLogger *log.Logger
var debugLogger *log.Logger

func containerRoutine(cli *client.Client, channel chan []types.Container) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			containers, _ := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Size: true})
			channel <- containers
		}
	}
}

func sendRoutine(mel *melody.Melody, channel chan []types.Container, urlPattern string) {
	for {
		containers := <-channel
		buff, err := json.Marshal(containers)
		if err != nil {
			fmt.Println(err)
		}
		filteredBroadCast(mel, buff, urlPattern)
	}
}

func filteredBroadCast(mel *melody.Melody, msg []byte, pattern string) {
	mel.BroadcastFilter(msg, func(session *melody.Session) bool {
		return session.Request.URL.Path == pattern
	})
}

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.37")
	r := gin.Default()
	m := melody.New()
	dashboardURL := "/dashboard"
	containerURL := "/container/:id"
	os.Mkdir("logs", os.ModePerm)
	ginLogFile, _ := os.Create("logs/gin.log")
	goLogFile, _ := os.Create("logs/go.log")

	gin.DefaultWriter = io.MultiWriter(ginLogFile)
	infoLogger = log.New(io.MultiWriter(goLogFile), "[INFO] : ", log.Lshortfile)
	debugLogger = log.New(io.MultiWriter(goLogFile), "[DEBUG] : ", log.Lshortfile)

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		panic(err)
	}

	containerChan := make(chan []types.Container)
	go containerRoutine(cli, containerChan)
	go sendRoutine(m, containerChan, dashboardURL+"/WS")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	r.GET(dashboardURL, func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "pages/dashboard.html")
	})

	r.Use(static.Serve("/public", static.LocalFile("./public", false)))

	r.GET(dashboardURL+"/WS", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	r.GET(containerURL, func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "pages/container.html")
	})

	r.GET(containerURL+"/WS", func(c *gin.Context) {
		id := c.Param("id")
		if val, ok := containers[id]; !ok {
			infoLogger.Printf("Channel %s not found. Adding new containers\n", id)
			cont := ContainerBroadcaster{subCount: 1, mel: m, logger: debugLogger, cli: cli}
			cont.Start(c.Request.URL.Path)
			containers[id] = &cont
		} else {
			val.Subscribe()
			containers[id] = val
		}
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleDisconnect(func(session *melody.Session) {
		if dashboardURL != session.Request.URL.Path {
			splittedUrl := strings.Split(session.Request.URL.Path, "/")
			id := splittedUrl[len(splittedUrl)-2]
			if val, ok := containers[id]; ok {
				val.Unsubscribe()
				containers[id] = val
				if val.subCount <= 0 {
					infoLogger.Printf("Channel %s prepared to close\n", id)
					val.Stop()
					delete(containers, id)
				}
			}
		}
	})

	r.Run(":3000")
}
