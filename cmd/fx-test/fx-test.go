package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// https://web-services.oanda.com/rates/api/v1/rates/USD.json?api_key=Le62T5rxE6Sewgv6MP1mNkFM&quote=AUD&quote=GBP&quote=NZD&fields=all&data_set=OANDA&decimal_places=5
var url string = "https://web-services.oanda.com/rates/api/v1/rates/"
var sampleResponse = "{\"base_currency\":\"USD\",\"meta\":{\"effective_params\":{\"data_set\":\"OANDA\",\"date\":\"2022-04-20\",\"decimal_places\":\"5\",\"fields\":[\"averages\",\"highs\",\"lows\",\"midpoint\"],\"quote_currencies\":[\"AUD\",\"GBP\",\"NZD\"]},\"request_time\":\"2022-04-20T01:58:29+0000\",\"skipped_currencies\":[]},\"quotes\":{\"AUD\":{\"ask\":\"1.35650\",\"bid\":\"1.35622\",\"date\":\"2022-04-19T23:59:59+0000\",\"high_ask\":\"1.36154\",\"high_bid\":\"1.36127\",\"low_ask\":\"1.35152\",\"low_bid\":\"1.35124\",\"midpoint\":\"1.35636\"},\"GBP\":{\"ask\":\"0.76894\",\"bid\":\"0.76881\",\"date\":\"2022-04-19T23:59:59+0000\",\"high_ask\":\"0.77038\",\"high_bid\":\"0.77027\",\"low_ask\":\"0.76688\",\"low_bid\":\"0.76675\",\"midpoint\":\"0.76888\"},\"NZD\":{\"ask\":\"1.48490\",\"bid\":\"1.48443\",\"date\":\"2022-04-19T23:59:59+0000\",\"high_ask\":\"1.48827\",\"high_bid\":\"1.48787\",\"low_ask\":\"1.47857\",\"low_bid\":\"1.47815\",\"midpoint\":\"1.48466\"}}}"
var apiKey = "Le62T5rxE6Sewgv6MP1mNkFM"

/// Function to construct the API request to retrieve market pricing for a supplied list of currencies.
func constructRequest(url string, apiKey string, baseCurrency string, quotes []string, decimalPlaces int) string {

	// Request format: <url>?:api_key=<api key>[&quote=<quotes>]+&fields=all&dataset=OANDA&decimal_places=<int>
	//var request string = url + "/AUD.json" + "?api_key=" + apiKey
	request := fmt.Sprintf("%s%s.json?api_key=%s", url, baseCurrency, apiKey)

	for _, q := range quotes {
		request += "&quote=" + q
	}
	request += "&fields=all&data_set=OANDA" // These are hard coded
	request += fmt.Sprintf("&decimal_places=%d", decimalPlaces)

	fmt.Printf("constructRequest(): API Request: %s\n", request)

	return request
}

func quoteGetter(c chan string, quit chan int) {

	msg := constructRequest(url, apiKey, "USD", []string{"AUD", "GBP", "NZD"}, 5)
	ticker := time.NewTicker(15 * time.Second)
	for _ = range ticker.C {
		resp := getPricing(msg)

		select {
		case c <- resp: // Send the pricing info to the main loop via the pricing channel.
			continue
		case <-quit: // Check if a quit signal has been received. If so, tell the main loop that all thread-termination steps are done..
			fmt.Printf("Received QUIT signal.\n")
			c <- "done"
			return

		}
	}
	fmt.Printf("ERROR: quoteGetter() exiting the thread incorrectly")
}

func getPricing(request string) string {
	return sampleResponse // Avoid smashing the free tier

	response, err := http.Get(request)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(string(responseData))

	return string(responseData)
}

func main() {
	fmt.Println("Initialising...")

	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	c := make(chan string) // Channel for passing pricing information
	quit := make(chan int) // Channel for sending quit signals.
	defer close(sigchan)
	defer close(c)
	defer close(quit)

	fmt.Printf("Starting\n")
	go quoteGetter(c, quit) // Start another thread to handle the comms

	// Process messages
	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			m := <-c // Test the channel to see if the price getter has retrieved a quote
			if m != "" {
				fmt.Printf("%s\n", m)
			}
		}
	}
	quit <- 0 // Send a quit signal

	// Wait for clean termination response from the thread.
	for q := <-c; q != "done"; {
		continue
	}
	fmt.Printf("Received clean termination signal from all threads.\n")
	fmt.Printf("Exiting")
}
