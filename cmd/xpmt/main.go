package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("error: command required\n")
		fmt.Printf("Available commands:\n" +
			"  fetch\n")
		return
	}
	switch os.Args[1] {
	case "fetch":
		fetch()
	default:
		fmt.Printf("error: unknown sub-command '%v'", os.Args[1])
	}
}

func fetch() {
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	apiKey := fetchCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	userId := fetchCmd.String("i", "", "User id to fetch variants for.")
	deviceId := fetchCmd.String("d", "", "Device id to fetch variants for.")
	userJson := fetchCmd.String("u", "", "The full user object to fetch variants for.")
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

	client := experiment.Initialize(*apiKey, &experiment.Config{Debug: *debug})
	variants, err := client.Fetch(user)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
	b, _ := json.Marshal(variants)
	fmt.Printf("%v\n", string(b))
}
