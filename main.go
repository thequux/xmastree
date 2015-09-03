package main

import (
	"net"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"time"
)

func echo(port int, conn net.Conn) {
	defer conn.Close()
	defer func() { connChan <- false }()
	connChan <- true
	io.Copy(conn, conn)
}

func listener(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(log.Fields{
			"port": port,
			"err": err,
		}).Error("Failed to listen")
		return
	}
	failureCount := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{"port": port, "err": err}).Warn("Failed to accept")
			failureCount++
			if failureCount > 5 {
				log.WithField("port", port).Error("Bailing")
				return
			}
			continue
		}
		failureCount = 0
		log.WithField("port", port).Debug("Connection received")
		go echo(port, conn)
	}
}

var connChan = make(chan bool, 20)

func conntrack() {
	// This is optimistic.
	var activeConnections uint64 = 0
	var totalConnections uint64 = 0
	tick := time.Tick(30*time.Second)
	for {
		select {
		case update := <-connChan:
			if update {
				activeConnections++
				totalConnections++
			} else {
				activeConnections--
			}
		case <-tick:
			log.WithFields(log.Fields{
				"active": activeConnections,
				"total": totalConnections,
			}).Info("Lies!")
		}
	}
}

func main() {
	log.SetLevel(log.DebugLevel)
	ret := make(chan int)
	go conntrack()

	for i := 1; i < 65536; i++ {
		go listener(i)
	}
	
	<-ret
}
