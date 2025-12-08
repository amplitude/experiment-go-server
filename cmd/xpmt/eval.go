package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/amplitude/experiment-go-server/pkg/experiment/local"
	"github.com/amplitude/experiment-go-server/pkg/experiment/remote"
)

func runEval(mgmtClient *managementClient) {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: xmpt eval [OPTIONS] [RAW_JSON]\n")
		os.Exit(1)
	}

	user := &experiment.User{}

	evalCmd := flag.NewFlagSet("eval", flag.ExitOnError)
	var deploymentName string
	evalCmd.StringVar(&deploymentName, "deployment-name", "", "Deployment name. If provided, need MANAGEMENT_API_KEY in env var to determine deployment-key")
	evalCmd.StringVar(&deploymentName, "n", "", "Deployment name. If provided, need MANAGEMENT_API_KEY in env var to determine deployment-key")
	var projectId string
	evalCmd.StringVar(&projectId, "project-id", "", "Project ID for selecting a deployment name. If provided, need MANAGEMENT_API_KEY in env var to determine deployment-key")
	var deploymentKey string
	evalCmd.StringVar(&deploymentKey, "deployment-key", "", "Deployment key. If not provided, read DEPLOYMENT_KEY from env var")
	evalCmd.StringVar(&deploymentKey, "k", "", "Deployment key. If not provided, read DEPLOYMENT_KEY from env var")
	var localMode bool
	evalCmd.BoolVar(&localMode, "local", false, "Use local evaluation")
	var remoteMode bool
	evalCmd.BoolVar(&remoteMode, "remote", false, "Use remote evaluation")
	var serverURL string
	evalCmd.StringVar(&serverURL, "server-url", "", "Server URL (overrides default)")

	flagKeys := newArrayFlags()
	evalCmd.Var(flagKeys, "flag-keys", "Flag keys (comma-separated or multiple -f flags)")
	evalCmd.Var(flagKeys, "f", "Flag keys (comma-separated or multiple -f flags)")

	evalCmd.StringVar(&user.UserId, "user-id", "", "User ID")
	evalCmd.StringVar(&user.UserId, "u", "", "User ID")
	evalCmd.StringVar(&user.DeviceId, "device-id", "", "Device ID")
	evalCmd.StringVar(&user.DeviceId, "d", "", "Device ID")

	userProps := newArrayFlags()
	evalCmd.Var(userProps, "user-props", "User properties (comma-separated prop1=val1,prop2=val2 or multiple -p flags)")
	evalCmd.Var(userProps, "p", "User properties (comma-separated prop1=val1,prop2=val2 or multiple -p flags)")

	cohortIds := newArrayFlags()
	evalCmd.Var(cohortIds, "cohort-ids", "Cohort IDs (comma-separated or multiple -c flags)")
	evalCmd.Var(cohortIds, "c", "Cohort IDs (comma-separated or multiple -c flags)")

	var groups string
	evalCmd.StringVar(&groups, "groups", "", "Groups JSON")

	var groupProps string
	evalCmd.StringVar(&groupProps, "group-props", "", "Group properties JSON")
	var groupCohortIds string
	evalCmd.StringVar(&groupCohortIds, "group-cohort-ids", "", "Group cohort IDs JSON")

	_ = parseFlags(evalCmd, os.Args[2:])

	props := userProps.toMap()
	user.UserProperties = make(map[string]interface{}, len(props))
	for k, v := range props {
		user.UserProperties[k] = v
	}

	user.CohortIds = make(map[string]struct{})
	for _, c := range cohortIds.toList() {
		user.CohortIds[c] = struct{}{}
	}
	if groups != "" {
		err := json.Unmarshal([]byte(groups), &user.Groups)
		if err != nil {
			fmt.Printf("error: invalid groups JSON: %v\n", err)
			os.Exit(1)
		}
	}
	if groupProps != "" {
		err := json.Unmarshal([]byte(groupProps), &user.GroupProperties)
		if err != nil {
			fmt.Printf("error: invalid group properties JSON: %v\n", err)
			os.Exit(1)
		}
	}
	if groupCohortIds != "" {
		err := json.Unmarshal([]byte(groupCohortIds), &user.GroupCohortIds)
		if err != nil {
			fmt.Printf("error: invalid group cohort IDs JSON: %v\n", err)
			os.Exit(1)
		}
	}

	var rawJSON string
	if len(evalCmd.Args()) > 0 {
		rawJSON = evalCmd.Args()[0]
	}

	if rawJSON != "" {
		err := json.Unmarshal([]byte(rawJSON), user)
		if err != nil {
			fmt.Printf("error: invalid JSON: %v\n", err)
			os.Exit(1)
		}
	}

	key := deploymentKey
	if key == "" {
		if deploymentName != "" {
			mgmtKey := os.Getenv("MANAGEMENT_API_KEY")
			if mgmtKey == "" {
				fmt.Printf("error: MANAGEMENT_API_KEY required when using -d\n")
				os.Exit(1)
			}
			ctx := context.Background()
			deployments, err := mgmtClient.getDeploymentByProjectIdLabel(ctx, &projectId, &deploymentName)
			if err != nil {
				fmt.Printf("error: failed to get deployment: %v\n", err)
				os.Exit(1)
			}
			if len(deployments) == 0 {
				fmt.Printf("no deployment found\n")
				os.Exit(1)
			}
			if len(deployments) > 1 {
				b, _ := json.Marshal(deployments[0])
				fmt.Printf("multiple deployments found, use first one: %v\n", string(b))
			}
			key = *deployments[0].Key
		} else {
			key = os.Getenv("DEPLOYMENT_KEY")
			if key == "" {
				fmt.Printf("error: deployment key required (use -k or set DEPLOYMENT_KEY)\n")
				os.Exit(1)
			}
		}
	}

	// Perform evaluation
	if remoteMode || (!localMode && !remoteMode) {
		// Perform remote evaluation
		config := &remote.Config{
			FetchTimeout: 10 * time.Second,
			RetryBackoff: &remote.RetryBackoff{FetchRetries: 0},
		}
		if serverURL != "" {
			config.ServerUrl = serverURL
		}
		client := remote.Initialize(key, config)
		variants, err := client.FetchV2(user)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		b, _ := json.Marshal(variants)
		fmt.Printf("%s\n", string(b))
	} else {
		// Perform local evaluation with experiment-go-server
		config := &local.Config{}
		if serverURL != "" {
			config.ServerUrl = serverURL
		}
		client := local.Initialize(key, config)
		err := client.Start()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		variants, err := client.EvaluateV2(user, flagKeys.toList())
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		b, _ := json.Marshal(variants)
		fmt.Printf("%s\n", string(b))
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
	_ = parseFlags(fetchCmd, os.Args[2:])

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
		FetchTimeout: 10 * time.Second,
		RetryBackoff: &remote.RetryBackoff{FetchRetries: 0},
	}

	if *url != "" {
		config.ServerUrl = *url
	} else if *staging {
		config.ServerUrl = "https://skylab-api.staging.amplitude.com"
	}

	client := remote.Initialize(*apiKey, config)
	start := time.Now()
	variants, err := client.FetchV2(user)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
	duration := time.Since(start)
	fmt.Println(duration)
	b, _ := json.Marshal(variants)
	fmt.Printf("%v\n", string(b))
}

func flags() {
	rulesCmd := flag.NewFlagSet("flags", flag.ExitOnError)
	apiKey := rulesCmd.String("k", "", "Api key for authorization, or use EXPERIMENT_KEY env var.")
	url := rulesCmd.String("url", "", "The server url to use to fetch variants from.")
	debug := rulesCmd.Bool("debug", false, "Log additional debug output to std out.")
	staging := rulesCmd.Bool("staging", false, "Use skylab staging environment.")
	_ = parseFlags(rulesCmd, os.Args[2:])

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
	flags, err := client.FlagsV2()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
	println(flags)
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
	_ = parseFlags(evaluateCmd, os.Args[2:])

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

	variants, err := client.EvaluateV2(user, nil)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}

	b, _ := json.Marshal(variants)
	fmt.Printf("%v\n", string(b))
}
