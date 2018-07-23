package main

import (
	"gopkg.in/olahol/melody.v1"
	"strings"
	"time"
)

type SubscribedChan struct {
	subCount int
	run      chan bool
}

func (sc SubscribedChan) Start(mel *melody.Melody, urlPattern string) {
	splittedUrl := strings.Split(urlPattern, "/")
	id := splittedUrl[len(splittedUrl)-2]
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if ok := isContainerPublisher(id); ok {
				//todo: here we need to call function to get new information.
				// And try to make it more usable by changin type to interface{}
				// And then call it to get all containers and specific one by creating one structure.
			} else {
				return
			}

		case _ = <-sc.run:
			return
		}
	}
}

func (sc SubscribedChan) Stop() {
	sc.run <- false
}
