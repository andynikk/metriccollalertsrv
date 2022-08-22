package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/andynikk/metriccollalertsrv/internal/consts"
	"github.com/andynikk/metriccollalertsrv/internal/handlers"
	"github.com/caarlos0/env/v6"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	go handleSignals(cancel)

	rp := handlers.NewRepStore()
	go http.ListenAndServe(consts.PortServer, rp.Router)

	cfg := &handlers.Config{}
	err := env.Parse(cfg)

	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	res := cfg.RESTORE
	patch := cfg.STORE_FILE

	if res {
		file, err := os.OpenFile(patch, os.O_RDONLY, 0777)
		if err != nil {
			fmt.Println("file not found")
		} else {
			ra, err := ioutil.ReadAll(file)
			if err != nil {
				fmt.Println("error read file")
			}
			fmt.Println(ra)
		}
	}

	for {
		select {
		case <-ctx.Done():
			arr := handlers.JSONMetricsAndValue(rp.MutexRepo)
			json_arr, err := json.Marshal(arr)
			if err != nil {
				panic(err)
			}

			fmt.Println(arr)

			//file, err := os.OpenFile(patch, os.O_CREATE|os.O_RDWR, 0777)
			file, err := os.Open(patch)
			if err != nil {
				panic(err)
			}
			io.WriteString(file, string(json_arr))
			file.Close()

			log.Panicln("server stopped")
			return
		default:

			timer := time.NewTimer(2 * time.Second)
			<-timer.C
		}
	}

}

func handleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)

	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			fmt.Println("canceled")
			cancel()
			return
		}
	}
}
