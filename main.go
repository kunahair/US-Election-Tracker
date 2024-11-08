package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// Define the structure for election results
type ElectionResults struct {
	Candidate1 string `json:"candidate1"`
	Candidate2 string `json:"candidate2"`
	Votes1     int    `json:"votes1"`
	Votes2     int    `json:"votes2"`
}

var lastResults ElectionResults

func main() {

	// // Set up the interval for checking updates
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// // Initialize the pastResults
	pastResults, _ := fetchElectionResults(0)
	printResults(pastResults)

	for {
		select {
		case <-ticker.C:
			// checkForUpdates()
			results, _ := fetchElectionResults(0)
			if results != pastResults {
				pastResults = results
				printResults(results)
			}
		}
	}
}

func printResults(results ElectionResults) {
	fmt.Println("Result changed at:", time.Now().Format(time.RFC1123))
	println(results.Candidate1 + " has " + fmt.Sprint(results.Votes1))
	println(results.Candidate2 + " has " + fmt.Sprint(results.Votes2))
	playChime()
}

func playChime() {
	// Open the chime sound file
	f, err := os.Open("chime.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Decode the sound file
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	// Initialize the speaker with the sample rate and buffer size
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Play the sound
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		fmt.Println("Chime finished playing")
	})))

	// Keep the program running until the sound finishes playing
	select {
	case <-time.After(time.Second * 5): // Adjust the duration as needed
	}
}

func fetchElectionResults(overrideValue int) (ElectionResults, error) {
	// Example URL for fetching results
	url := "https://www.theguardian.com/us-news/us-elections-2024"

	// Fetch the HTML content
	resp, err := http.Get(url)
	if err != nil {
		return ElectionResults{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ElectionResults{}, fmt.Errorf("failed to fetch page: %s", resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ElectionResults{}, err
	}

	// Parse the HTML and extract election results
	results := ElectionResults{}

	harrisVotes := doc.Find("#gv-atom-us-election-tracker-2024 > div > div > div.results.svelte-ydurd5 > div.candidates.svelte-ydurd5 > div.candidate.dem.svelte-ydurd5 > div.candidate-summary.svelte-ydurd5 > span.electoral-college-votes-count.svelte-ydurd5").Text()
	trumpVotes := doc.Find("#gv-atom-us-election-tracker-2024 > div > div > div.results.svelte-ydurd5 > div.candidates.svelte-ydurd5 > div.candidate.gops.svelte-ydurd5 > div.candidate-summary.svelte-ydurd5 > span.electoral-college-votes-count.svelte-ydurd5").Text()

	// Assuming the HTML structure has specific IDs or classes to locate the data
	results.Candidate1 = "Kamala Harris"
	results.Candidate2 = "Donald Trump"

	// Assuming the votes are inside elements with classes .votes1 and .votes2
	votes1 := harrisVotes
	votes2 := trumpVotes

	// Convert votes from string to int
	fmt.Sscanf(votes1, "%d", &results.Votes1)
	fmt.Sscanf(votes2, "%d", &results.Votes2)

	if overrideValue != 0 {
		results.Votes1 = overrideValue
	}

	return results, nil
}
