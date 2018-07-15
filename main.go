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
)

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
			channel <- container
		}
	}
}

func sendRoutine(melody *melody.Melody, channel chan []types.Container) {
	for {
		containers := <-channel
		buff, err := json.Marshal(containers)
		if err != nil {
			fmt.Println(err)
		}
		melody.Broadcast(buff)
	}
}

func sendJSONContainerRoutine(melody *melody.Melody, channel chan types.ContainerJSON) {
	for {
		containers := <-channel
		buff, err := json.Marshal(containers)
		if err != nil {
			fmt.Println(err)
		}
		melody.Broadcast(buff)
	}
}

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.37")
	r := gin.Default()
	m := melody.New()

	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println(err)
	}

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "pages/dashboard.html")
	})

	r.Use(static.Serve("/public", static.LocalFile("./public", false)))

	r.GET("/dashboardWS", func(c *gin.Context) {
		containerChan := make(chan []types.Container)
		go containerRoutine(cli, containerChan)
		go sendRoutine(m, containerChan)
		m.HandleRequest(c.Writer, c.Request)
	})

	r.GET("/container/:id", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "pages/container.html")
	})

	r.GET("/container/:id/WS", func(c *gin.Context) {
		id := c.Param("id")
		containerChan := make(chan types.ContainerJSON)
		go singleContainerRoutine(id, cli, containerChan)
		go sendJSONContainerRoutine(m, containerChan)
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleDisconnect(func(s *melody.Session) {
		fmt.Println("dummyDisco")
	})

	m.HandleConnect(func(s *melody.Session) {
		fmt.Println("dummyConnect")
	})

	r.Run(":3000")
}
