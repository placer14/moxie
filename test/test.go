package main

import (
	"log"
	"net/http"
)

var maxRequests = 5
var requests = make(chan bool, maxRequests)

func main() {
	for i := 0; ; i++ {
		go func() {
			requests <- true
			log.Println("Sending google...")
			resp, err := http.Get("http://google.com")
			if err != nil {
				log.Println("...fail", i, err.Error())
			} else {
				log.Println("...done google", i, resp.Status)
			}
			<-requests
		}()
		go func() {
			requests <- true
			log.Println("Sending reddit...")
			resp, err := http.Get("http://reddit.com")
			if err != nil {
				log.Println("...fail", i, err.Error())
			} else {
				log.Println("...done reddit", i, resp.Status)
			}
			<-requests
		}()
	}

}
