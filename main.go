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
	"sync"
	"time"
)

type SubscribedChan struct {
	subCount int
	channel  chan types.ContainerJSON
}

var containers = sync.Map{}
var infologger *log.Logger

func containerRoutine(cli *client.Client, channel chan []types.Container) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			containers, _ := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
			channel <- containers
		}
	}
}

func singleContainerRoutine(containerID string, cli *client.Client, channel chan types.ContainerJSON) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			container, _ := cli.ContainerInspect(context.Background(), containerID)
			if _, ok := containers.Load(containerID); ok {
				channel <- container
			} else {
				infologger.Printf("Closed routine for %s\n", containerID)
				close(channel)
				return
			}
		default:
			if _, ok := containers.Load(containerID); !ok {
				infologger.Printf("Closed routine for %s\n", containerID)
				close(channel)
				return
			}
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

func sendJSONContainerRoutine(mel *melody.Melody, channel chan types.ContainerJSON, urlPattern string) {
	splittedUrl := strings.Split(urlPattern, "/")
	id := splittedUrl[len(splittedUrl)-2]
	for {
		if _, ok := containers.Load(id); !ok {
			infologger.Printf("Stopped broadcasting for %s\n", id)
			return
		} else {
			containers := <-channel
			buff, err := json.Marshal(containers)
			if err != nil {
				fmt.Println(err)
			}
			filteredBroadCast(mel, buff, urlPattern)
		}
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
	infologger = log.New(io.MultiWriter(goLogFile), "INFO: ", log.Lshortfile)

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		fmt.Println(err)
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
		if val, ok := containers.Load(id); !ok {
			infologger.Printf("Channel %s not found. Adding new channel\n", id)
			container := make(chan types.ContainerJSON)
			go singleContainerRoutine(id, cli, container)
			go sendJSONContainerRoutine(m, container, c.Request.URL.Path)
			containers.Store(id, SubscribedChan{subCount: 1, channel: container})
		} else {
			value := val.(SubscribedChan)
			value.subCount = value.subCount + 1
			containers.Store(id, value)
		}
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleDisconnect(func(session *melody.Session) {
		if dashboardURL != session.Request.URL.Path {
			splittedUrl := strings.Split(session.Request.URL.Path, "/")
			id := splittedUrl[len(splittedUrl)-2]
			if val, ok := containers.Load(id); ok {
				value := val.(SubscribedChan)
				value.subCount = value.subCount - 1
				containers.Store(id, val)
				if value.subCount <= 0 {
					infologger.Printf("Channel %s prepared to close\n", id)
					containers.Delete(id)
				}
			}
		}
	})

	m.HandleConnect(func(session *melody.Session) {

	})

	r.Run(":3000")
}
