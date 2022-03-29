package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("error: command required\n")
		fmt.Printf("Available commands:\n" +
			"  fetch\n" +
			"  rules\n")
		return
	}
	switch os.Args[1] {
	case "fetch":
		fetch()
	case "rules":
		rules()
	case "stream":
		stream()
	case "publish":
		publish(os.Args)
	case "spam":
		spam()
	case "test":
		test()
	default:
		fmt.Printf("error: unknown sub-command '%v'", os.Args[1])
	}
}

// status, isTimeout, err
func testInternal() (int, bool) {
	client := &http.Client{}
	endpoint, err := url.Parse("https://akamai-test.lab.amplitude.com/")
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return -1, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.String(), nil)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return -1, false
	}
	req.Header.Set("id", randStringRunes(10))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return -1, false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
		return -1, false
	}
	if resp.StatusCode == http.StatusInternalServerError {
		if len(body) > 0 && body[0] == '<' {
			fmt.Printf("%v, TIMEOUT\n", resp.StatusCode)
			return resp.StatusCode, true // TIMEOUT
		} else {
			fmt.Printf("%v, internal\n", resp.StatusCode)
			return resp.StatusCode, false // INTERNAL
		}
	} else {
		fmt.Println(resp.StatusCode)
		return resp.StatusCode, false
	}

}

func test() {
	success := 0
	timeoutError := 0
	internalError := 0

	for i := 0; i < 1000; i++ {
		code, isTimeout := testInternal()
		if code == 200 {
			success += 1
		} else if code == 500 && isTimeout {
			timeoutError += 1
		} else if code == 500 && !isTimeout {
			internalError += 1
		}
	}
	fmt.Printf("Success: %v\nTimeout: %v\nInternal: %v\n", success, timeoutError, internalError)
}

func fetch() {
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	apiKey := fetchCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	userId := fetchCmd.String("i", "", "User id to fetch variants for.")
	deviceId := fetchCmd.String("d", "", "Device id to fetch variants for.")
	userJson := fetchCmd.String("u", "", "The full user object to fetch variants for.")
	url := fetchCmd.String("url", "", "The server url to use to fetch variants from.")
	debug := fetchCmd.Bool("debug", false, "Log additional debug output to std out.")
	staging := fetchCmd.Bool("staging", false, "Use skylab staging environment.")
	_ = fetchCmd.Parse(os.Args[2:])

	if len(os.Args) == 2 {
		fetchCmd.Usage()
		return
	}

	if apiKey == nil || *apiKey == "" {
		envKey := os.Getenv("EXPERIMENT_KEY")
		if envKey == "" {
			fetchCmd.Usage()
			fmt.Printf("error: must set experiment api key using cli flag or EXPERIMENT_KEY env var\n")
			os.Exit(1)
			return
		}
		apiKey = &envKey
	}

	user := &experiment.User{}
	if userJson != nil && *userJson != "" {
		err := json.Unmarshal([]byte(*userJson), user)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
			return
		}
	}
	if userId != nil && *userId != "" {
		user.UserId = *userId
	}
	if deviceId != nil && *deviceId != "" {
		user.DeviceId = *deviceId
	}

	config := &experiment.Config{
		Debug:        *debug,
		RetryBackoff: &experiment.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := experiment.Initialize(*apiKey, config)

	variants, err := client.Fetch(user)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
	b, _ := json.Marshal(variants)
	fmt.Printf("%v\n", string(b))
}

func rules() {
	rulesCmd := flag.NewFlagSet("rules", flag.ExitOnError)
	apiKey := rulesCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	url := rulesCmd.String("url", "", "The server url to use to fetch variants from.")
	debug := rulesCmd.Bool("debug", false, "Log additional debug output to std out.")
	staging := rulesCmd.Bool("staging", false, "Use skylab staging environment.")
	_ = rulesCmd.Parse(os.Args[2:])

	if len(os.Args) == 3 && os.Args[1] == "--help" {
		rulesCmd.Usage()
		return
	}

	if apiKey == nil || *apiKey == "" {
		envKey := os.Getenv("EXPERIMENT_KEY")
		if envKey == "" {
			rulesCmd.Usage()
			fmt.Printf("error: must set experiment api key using cli flag or EXPERIMENT_KEY env var\n")
			os.Exit(1)
			return
		}
		apiKey = &envKey
	}

	config := &experiment.Config{
		Debug:        *debug,
		RetryBackoff: &experiment.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := experiment.Initialize(*apiKey, config)
	variants, err := client.Rules()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
	b, _ := json.Marshal(variants)
	st, _ := strconv.Unquote(string(b))
	fmt.Println(st)
}

func stream() {
	fetchCmd := flag.NewFlagSet("stream", flag.ExitOnError)
	apiKey := fetchCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	userId := fetchCmd.String("i", "", "User id to fetch variants for.")
	deviceId := fetchCmd.String("d", "", "Device id to fetch variants for.")
	userJson := fetchCmd.String("u", "", "The full user object to fetch variants for.")
	url := fetchCmd.String("url", "", "The server url to use to fetch variants from.")
	debug := fetchCmd.Bool("debug", false, "Log additional debug output to std out.")
	_ = fetchCmd.Parse(os.Args[2:])

	if len(os.Args) == 2 {
		fetchCmd.Usage()
		return
	}

	if apiKey == nil || *apiKey == "" {
		envKey := os.Getenv("EXPERIMENT_KEY")
		if envKey == "" {
			fetchCmd.Usage()
			fmt.Printf("error: must set experiment api key using cli flag or EXPERIMENT_KEY env var\n")
			os.Exit(1)
			return
		}
		apiKey = &envKey
	}

	user := &experiment.User{}
	if userJson != nil && *userJson != "" {
		err := json.Unmarshal([]byte(*userJson), user)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
			return
		}
	}
	if userId != nil && *userId != "" {
		user.UserId = *userId
	}
	if deviceId != nil && *deviceId != "" {
		user.DeviceId = *deviceId
	}

	config := &experiment.Config{
		Debug:        *debug,
		RetryBackoff: &experiment.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	}

	client := experiment.Initialize(*apiKey, config)
	variantsChannel, err := client.Stream(user)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
	for {
		select {
		case variants, ok := <-variantsChannel:
			if !ok {
				return
			}
			variantsJson, _ := json.Marshal(variants)
			fmt.Printf("%v\n\n", string(variantsJson))
		}
	}
}

func publish(args []string) {
	fetchCmd := flag.NewFlagSet("publish", flag.ExitOnError)
	apiKey := fetchCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	url := fetchCmd.String("url", "", "The server url to use to fetch variants from.")
	debug := fetchCmd.Bool("debug", false, "Log additional debug output to std out.")
	_ = fetchCmd.Parse(os.Args[3:])

	if apiKey == nil || *apiKey == "" {
		envKey := os.Getenv("EXPERIMENT_KEY")
		if envKey == "" {
			fetchCmd.Usage()
			fmt.Printf("error: must set experiment api key using cli flag or EXPERIMENT_KEY env var\n")
			os.Exit(1)
			return
		}
		apiKey = &envKey
	}

	config := &experiment.Config{
		Debug:        *debug,
		RetryBackoff: &experiment.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	}

	client := experiment.Initialize(*apiKey, config)
	err := client.Publish(args[2], "")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func spam() {
	fetchCmd := flag.NewFlagSet("spam", flag.ExitOnError)
	apiKey := fetchCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	userId := fetchCmd.String("i", "", "User id to fetch variants for.")
	deviceId := fetchCmd.String("d", "", "Device id to fetch variants for.")
	userJson := fetchCmd.String("u", "", "The full user object to fetch variants for.")
	url := fetchCmd.String("url", "", "The server url to use to fetch variants from.")
	debug := fetchCmd.Bool("debug", false, "Log additional debug output to std out.")
	staging := fetchCmd.Bool("staging", false, "Use skylab staging environment.")
	numRequests := fetchCmd.Int("n", 300, "The number of requests to send.")
	maxConcurrency := fetchCmd.Int("c", 10, "The number of concurrent requests.")
	_ = fetchCmd.Parse(os.Args[2:])

	if len(os.Args) == 2 {
		fetchCmd.Usage()
		return
	}

	if apiKey == nil || *apiKey == "" {
		envKey := os.Getenv("EXPERIMENT_KEY")
		if envKey == "" {
			fetchCmd.Usage()
			fmt.Printf("error: must set experiment api key using cli flag or EXPERIMENT_KEY env var\n")
			os.Exit(1)
			return
		}
		apiKey = &envKey
	}

	user := &experiment.User{}
	if userJson != nil && *userJson != "" {
		err := json.Unmarshal([]byte(*userJson), user)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
			return
		}
	}
	if userId != nil && *userId != "" {
		user.UserId = *userId
	}
	if deviceId != nil && *deviceId != "" {
		user.DeviceId = *deviceId
	}

	config := &experiment.Config{
		Debug:        *debug,
		RetryBackoff: &experiment.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := experiment.Initialize(*apiKey, config)

	ctx := context.Background()
	sem := semaphore.NewWeighted(int64(*maxConcurrency))
	var group sync.WaitGroup
	for i := 0; i < *numRequests; i++ {
		group.Add(1)
		go func() {
			_ = sem.Acquire(ctx, 1)
			defer group.Done()
			defer sem.Release(1)
			_, _ = client.Fetch(user)
		}()
	}
	group.Wait()
}

// HELPER

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
