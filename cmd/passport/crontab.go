package main

import (
	"github.com/robfig/cron"
	"log"
)

func main() {
	log.Println("Starting...")

	c := cron.New()
	c.AddFunc(""+
		"0 * * * * *", func() {
		log.Println("Run Job1...")
	})
	c.AddFunc("* * * * * *", func() {
		log.Println("Run Job2...")
	})

	c.Start()

	select {}
	/*
		t1 := time.NewTimer(time.Second * 10)
		for {
			select {
			case <-t1.C:
				t1.Reset(time.Second * 10)
			}
		}

	*/
}
