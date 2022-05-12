package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/amplitude/experiment-go-server/pkg/experiment/local"
	"github.com/amplitude/experiment-go-server/pkg/experiment/remote"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("error: command required\n")
		fmt.Printf("Available commands:\n" +
			"  fetch\n" +
			"  rules\n" +
			"  evaluate\n")
		return
	}
	switch os.Args[1] {
	case "fetch":
		fetch()
	case "rules":
		rules()
	case "evaluate":
		evaluate()
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

	config := &remote.Config{
		Debug:        *debug,
		RetryBackoff: &remote.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := remote.Initialize(*apiKey, config)

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

	config := &local.Config{
		Debug: *debug,
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := local.Initialize(*apiKey, config)
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

func evaluate() {
	evaluateCmd := flag.NewFlagSet("evaluate", flag.ExitOnError)
	apiKey := evaluateCmd.String("k", "", "Server api key for authorization, or use EXPERIMENT_KEY env var.")
	userId := evaluateCmd.String("i", "", "User id to fetch variants for.")
	deviceId := evaluateCmd.String("d", "", "Device id to fetch variants for.")
	userJson := evaluateCmd.String("u", "", "The full user object to fetch variants for.")
	url := evaluateCmd.String("url", "", "The server url to use poll for flag configs from.")
	debug := evaluateCmd.Bool("debug", false, "Log additional debug output to std out.")
	staging := evaluateCmd.Bool("staging", false, "Use skylab staging environment.")
	_ = evaluateCmd.Parse(os.Args[2:])

	if len(os.Args) == 2 {
		evaluateCmd.Usage()
		return
	}

	if apiKey == nil || *apiKey == "" {
		envKey := os.Getenv("EXPERIMENT_KEY")
		if envKey == "" {
			evaluateCmd.Usage()
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

	config := &local.Config{
		Debug: *debug,
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := local.Initialize(*apiKey, config)
	err := client.Start()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	//for {
	//	time.Sleep(time.Duration(rand.Intn(1000)) * time.Microsecond)
	//	start := time.Now()
	//
	//	_, err := client.Evaluate(user, nil)
	//	if err != nil {
	//		fmt.Printf("error: %v\n", err)
	//		os.Exit(1)
	//		return
	//	}
	//
	//	duration := time.Since(start)
	//	fmt.Println(duration)
	//}

	variants, err := client.Evaluate(user, nil)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}

	b, _ := json.Marshal(variants)
	fmt.Printf("%v\n", string(b))
}
