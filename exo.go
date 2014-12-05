package main

import (
	"egoscale"
	"os"
	"fmt"
)

func main() {

	endpoint := os.Getenv("EXOSCALE_ENDPOINT")
	apiKey := os.Getenv("EXOSCALE_API_KEY")
	apiSecret:= os.Getenv("EXOSCALE_API_SECRET")
	client := egoscale.NewClient(endpoint, apiKey, apiSecret)

	topo, err := client.GetTopology()
	if err != nil {
		fmt.Printf("got error: %v\n", err)
		return
	}
	fmt.Printf("got response: %+v\n", topo)
}
