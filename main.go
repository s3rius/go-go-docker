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

func sendRoutine(mel *melody.Melody, channel chan []types.Container, urlPattern string) {
	for {
		containers := <-channel
		buff, err := json.Marshal(containers)
		if err != nil {
			fmt.Println(err)
		}
		mel.BroadcastFilter(buff, func(session *melody.Session) bool {
			return session.Request.URL.Path == urlPattern
		})
	}
}

func sendJSONContainerRoutine(mel *melody.Melody, channel chan types.ContainerJSON, urlPattern string) {
	for {
		containers := <-channel
		buff, err := json.Marshal(containers)
		if err != nil {
			fmt.Println(err)
		}
		mel.BroadcastFilter(buff, func(session *melody.Session) bool {
			return session.Request.URL.Path == urlPattern
		})
		//mel.Broadcast(buff)
	}
}

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.37")
	r := gin.Default()
	m := melody.New()

	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		fmt.Println(err)
	}



	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	r.GET("/dashboard", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "pages/dashboard.html")
	})

	r.Use(static.Serve("/public", static.LocalFile("./public", false)))

	r.GET("/dashboard/dashboardWS", func(c *gin.Context) {
		containerChan := make(chan []types.Container)
		go containerRoutine(cli, containerChan)
		go sendRoutine(m, containerChan, c.Request.URL.Path)
		m.HandleRequest(c.Writer, c.Request)
	})

	r.GET("/container/:id", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "pages/container.html")
	})

	r.GET("/container/:id/WS", func(c *gin.Context) {
		id := c.Param("id")
		container := make(chan types.ContainerJSON)
		go singleContainerRoutine(id, cli, container)
		go sendJSONContainerRoutine(m, container, c.Request.URL.Path)
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
