// Simple CMD CLI for outputting the data every 30 secs (data still comes every 61secs)
package main

import (
	"flag"
	"fmt"
	"time"

	d2 "kimpton.io/dclonetracker/d2ioscraper"
)

func main(){

	flag.Parse()

	fmt.Printf("Running...\n")

	// Waiting for some state
	var prevState []d2.D2CloneState
	for len(prevState) == 0 {
		time.Sleep(5 * time.Second)
		prevState = d2.GetGlobalState()
	}

	// Print initial list
	fmt.Printf("Initial List")
	for _, d := range prevState {
		fmt.Printf("%s\n", d)
	}

	// Loop forever and print updates
	for {
		
		state := d2.GetGlobalState()
		// Print the results
		for _, d := range state {
			for _, pd := range prevState {
				if updated, err := d.HasUpdated(pd); err == nil && updated {
					fmt.Printf("[Update] %s\n", d)
				}
			}
			
		}
		fmt.Printf("\n")
		prevState = state
		time.Sleep(30 * time.Second)
	}
	
}