package main

import (
	"context"
	"github.com/docker/docker/client"
	"gopkg.in/olahol/melody.v1"
	"log"
	"strings"
	"time"
)

type ContainerBroadcaster struct {
	subCount int
	logger   *log.Logger
	mel      *melody.Melody
	cli      *client.Client
	quit     chan bool
}

func (cb *ContainerBroadcaster) Start(urlPattern string) {
	splittedUrl := strings.Split(urlPattern, "/")
	id := splittedUrl[len(splittedUrl)-2]
	cb.quit = make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-cb.quit:
				cb.logger.Println("Stop signal received. Stopping...")
				return
			case <-ticker.C:
				_, raw, err := cb.cli.ContainerInspectWithRaw(context.Background(), id, true)
				cb.logger.Println("Pushing new info")
				if err != nil {
					cb.logger.Printf("Error occured while updating info. Closing %s\n", id)
					cb.logger.Printf("err: %s", err)
					return
				}
				cb.filteredBroadcast(raw, urlPattern)
			}
		}
	}()
}

func (cb *ContainerBroadcaster) filteredBroadcast(msg []byte, pattern string) {
	cb.mel.BroadcastFilter(msg, func(session *melody.Session) bool {
		return session.Request.URL.Path == pattern
	})
}

func (cb *ContainerBroadcaster) SubCount() int {
	return cb.subCount
}

func (cb *ContainerBroadcaster) Unsubscribe() {
	cb.subCount = cb.subCount - 1
}

func (cb *ContainerBroadcaster) Subscribe() {
	cb.subCount = cb.subCount + 1
}

func (cb *ContainerBroadcaster) Stop() {
	cb.quit <- true
}
