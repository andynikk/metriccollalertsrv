package main

import (
	"fmt"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	fmt.Println("+++++++++++++++ start agent", 1)
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	fmt.Println("+++++++++++++++ start agent", 2)
	//config := environment.InitConfigAgent()
	//fmt.Println("+++++++++++++++ start agent", 3)
	//a := agent.NewAgent(config)
	//fmt.Println("+++++++++++++++ start agent", 4)
	//a.Run()
	//fmt.Println("+++++++++++++++ start agent", 5)
	//
	//stop := make(chan os.Signal, 1)
	//fmt.Println("+++++++++++++++ start agent", 6)
	//signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	//fmt.Println("+++++++++++++++ start agent", 7)
	//<-stop
	//fmt.Println("+++++++++++++++ start agent", 8)
	//
	//a.Stop()
}
