package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"gopkg.in/olahol/melody.v1"
	"strings"
)

type Pair struct {
	first  int
	second chan types.ContainerJSON
}

var containers = make(map[string]Pair)

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
			if _, ok := containers[containerID]; ok{
				channel <- container
			}else {
				fmt.Println("Closed routine for " + containerID)
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
		if val, ok := containers[id]; !ok {
			container := make(chan types.ContainerJSON)
			go singleContainerRoutine(id, cli, container)
			go sendJSONContainerRoutine(m, container, c.Request.URL.Path)
			containers[id] = Pair{first: 1, second: container}
		} else {
			val.first = val.first + 1
			containers[id] = val
		}
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleDisconnect(func(session *melody.Session) {
		splittedUrl := strings.Split(session.Request.URL.Path, "/")
		id := splittedUrl[len(splittedUrl)-2]
		if val, ok := containers[id]; ok {
			val.first = val.first - 1
			containers[id] = val
			if val.first <= 0 {
				delete(containers, id)
			}
		}
	})

	r.Run(":3000")
}
