// This package focuses on retrieving data from diablo2.io using their d2clone tracker api https://diablo2.io/dclone_api.php
package d2ioscraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const(
	MIN_POLL_DELAY = 61 * time.Second // Time in seconds || ToS of API is 60 seconds between requests
	ENDPOINT = "https://diablo2.io/dclone_api.php"
)

// Start the scrape loop
func init(){
	go scrapeLoop()
}

// Scrape loop for getting the data and updating global state
func scrapeLoop() {
	for {
		// Wait first incase previous instance polled recently
		time.Sleep(MIN_POLL_DELAY)

		// Get new data & parse
		state, err := scrapOnce()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue // Go to next loop
		}

		// Update the global state
		Update(state)
	}
}

func scrapOnce() ([]D2CloneState, error) {

	// Make a http request
	resp, err := http.Get(ENDPOINT)
	// Check there was no error with request
	if err != nil {
		return []D2CloneState{}, err
	}
	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return []D2CloneState{}, fmt.Errorf("url %s returned status code %d", ENDPOINT, resp.StatusCode)
	}
	
	// Read the body & handle error
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []D2CloneState{}, err
	}
	return parseResponse(body)
}

// Takes the body from https://diablo2.io/dclone_api.php and parses it to []D2CloneState
func parseResponse(body []byte) ([]D2CloneState, error) {
	// The page stores data in an array of grouped key/value pairs
	var data []map[string]string
	// Unmarshal the body into the var above
	json.Unmarshal(body, &data)

	state := []D2CloneState{}

	for _, d := range data {

		newState := D2CloneState{}

		// Region Parser
		if val, ok := d["region"]; ok {
			region, err := strconv.Atoi(val)
			if err != nil {
				// If no region, print issue and go to next case
				fmt.Printf("unknown region %s\n", val)
				continue
			}
			newState.Region = Region(region)
		}

		// Ladder Parser
		if val, ok := d["ladder"]; ok {
			if val == "1" {
				// Ladder
				newState.IsLadder = true
			} else if val == "2" {
				// Non Ladder
				newState.IsLadder = false
			} else {
				// If no valid ladder info, print issue and go to next case
				fmt.Printf("unknown ladder setting %s\n", val)
				continue
			}
		}

		// Hardcore parser
		if val, ok := d["hc"]; ok {
			if val == "1" {
				// Hardcore
				newState.IsHardcore = true
			} else if val == "2" {
				// Softcore
				newState.IsHardcore = false
			} else {
				// If no valid hc info, print issue and go to next case
				fmt.Printf("unknown hardcore setting %s\n", val)
				continue
			}
		}

		// Progress parser
		if val, ok := d["progress"]; ok {
			progress, err := strconv.Atoi(val)
			if err != nil || progress < 0 || progress > MAX_CLONE_LEVELS {
				// If no progress, print issue and go to next case
				fmt.Printf("unknown progress %s\n", val)
				continue
			}
			newState.Progress = uint64(progress)
		}

		// Timestamp parser
		if val, ok := d["timestamped"]; ok {
			unixtime, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				// If no progress, print issue and go to next case
				fmt.Printf("unknown timestamp %s\n", val)
				continue
			}
			newState.LastUpdated = time.Unix(unixtime, 0)
		}

		// Append to list
		state = append(state, newState)
	}

	return state, nil
}