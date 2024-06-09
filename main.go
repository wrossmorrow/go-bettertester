package main

import (
	"flag"

	bt "github.com/wrossmorrow/bettertester"
)

var (
	filename = flag.String("config", "./examples/example.yaml", "config file")
)

func main() {
	flag.Parse()
	callgraph := bt.NewCallGraphFromConfigFile(*filename)
	ctx := bt.NewExecutionContext()
	err := callgraph.Execute(ctx)
	if err != nil {
		panic(err)
	}

}
