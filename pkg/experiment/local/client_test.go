package local

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/joho/godotenv"
)

var client *Client

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	projectApiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	cohortSyncConfig := CohortSyncConfig{
		ApiKey:    projectApiKey,
		SecretKey: secretKey,
	}
	client = Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz",
		&Config{CohortSyncConfig: &cohortSyncConfig})
	err = client.Start()
	if err != nil {
		panic(err)
	}
}

func TestClientInitialize(t *testing.T) {
	client1 := Initialize("apiKey1", nil)
	client2 := Initialize("apiKey1", nil)
	client3 := Initialize("apiKey2", nil)
	if client1 != client2 {
		t.Fatalf("Expected equal client references.")
	}
	if client1 == client3 {
		t.Fatalf("Expected different client references.")
	}
}

func TestEvaluate(t *testing.T) {
	user := &experiment.User{UserId: "test_user"}
	result, err := client.Evaluate(user, nil)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-evaluation-ci-test"]
	if variant.Key != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Payload != "payload" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	variant = result["sdk-ci-test"]
	if variant.Key != "" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestEvaluateV2AllFlags(t *testing.T) {
	user := &experiment.User{UserId: "test_user"}
	result, err := client.EvaluateV2(user, nil)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-evaluation-ci-test"]
	if variant.Key != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Payload != "payload" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	variant = result["sdk-ci-test"]
	if variant.Key != "off" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestEvaluateV2OneFlag(t *testing.T) {
	user := &experiment.User{UserId: "test_user"}
	flagKeys := []string{"sdk-local-evaluation-ci-test"}
	result, err := client.EvaluateV2(user, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-evaluation-ci-test"]
	if variant.Key != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Payload != "payload" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestEvaluateV2AllFlagsWithDependencies(t *testing.T) {
	user := &experiment.User{UserId: "user_id", DeviceId: "device_id"}
	result, err := client.EvaluateV2(user, nil)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-ci-local-dependencies-test"]
	if variant.Key != "control" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "control" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestEvaluateV2OneFlagWithDependencies(t *testing.T) {
	user := &experiment.User{UserId: "user_id", DeviceId: "device_id"}
	flagKeys := []string{"sdk-ci-local-dependencies-test"}
	result, err := client.EvaluateV2(user, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-ci-local-dependencies-test"]
	if variant.Key != "control" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "control" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestEvaluateV2UnknownFlagKey(t *testing.T) {
	user := &experiment.User{UserId: "user_id", DeviceId: "device_id"}
	flagKeys := []string{"does-not-exist"}
	result, err := client.EvaluateV2(user, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-dependencies-test"]
	if variant.Key != "" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestFlagMetadataUnknownFlagKey(t *testing.T) {
	md := client.FlagMetadata("does-not-exist")
	if md != nil {
		t.Fatalf("Unexpected metadata %v", md)
	}
}

func TestFlagMetadataLocalFlagKey(t *testing.T) {
	md := client.FlagMetadata("sdk-local-evaluation-ci-test")
	if md["evaluationMode"] != "local" {
		t.Fatalf("Unexpected metadata %v", md)
	}
}

func TestEvaluateV2Cohort(t *testing.T) {
	targetedUser := &experiment.User{UserId: "12345"}
	nonTargetedUser := &experiment.User{UserId: "not_targeted"}
	flagKeys := []string{"sdk-local-evaluation-user-cohort-ci-test"}
	result, err := client.EvaluateV2(targetedUser, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-evaluation-user-cohort-ci-test"]
	if variant.Key != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	result, err = client.EvaluateV2(nonTargetedUser, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant = result["sdk-local-evaluation-user-cohort-ci-test"]
	if variant.Key != "off" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestEvaluateV2GroupCohort(t *testing.T) {
	targetedUser := &experiment.User{
		UserId:   "12345",
		DeviceId: "device_id",
		Groups: map[string][]string{
			"org id": {"1"},
		}}
	nonTargetedUser := &experiment.User{
		UserId:   "12345",
		DeviceId: "device_id",
		Groups: map[string][]string{
			"org id": {"not_targeted"},
		}}
	flagKeys := []string{"sdk-local-evaluation-group-cohort-ci-test"}
	result, err := client.EvaluateV2(targetedUser, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-evaluation-group-cohort-ci-test"]
	if variant.Key != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	result, err = client.EvaluateV2(nonTargetedUser, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant = result["sdk-local-evaluation-group-cohort-ci-test"]
	if variant.Key != "off" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}


func TestMain(t *testing.T) {
	client := Initialize("server-tUTqR62DZefq7c73zMpbIr1M5VDtwY8T", &Config{ServerUrl: "noserver", StreamUpdates: true, StreamServerUrl: "https://skylab-stream.stag2.amplitude.com"})
	client.Start()
	println(client.flagConfigStorage.getFlagConfigs(), len(client.flagConfigStorage.getFlagConfigs()))
	time.Sleep(2000 * time.Millisecond)
	println(client.flagConfigStorage.getFlagConfigs(), len(client.flagConfigStorage.getFlagConfigs()))

	// connTimeout := 1500 * time.Millisecond
	// api := NewFlagConfigStreamApiV2("server-tUTqR62DZefq7c73zMpbIr1M5VDtwY8T", "https://skylab-stream.stag2.amplitude.com", connTimeout)
	// cohortStorage := newInMemoryCohortStorage()
	// flagConfigStorage := newInMemoryFlagConfigStorage()
	// dr := newDeploymentRunner(
	// 	DefaultConfig, 
	// 	NewFlagConfigApiV2("server-tUTqR62DZefq7c73zMpbIr1M5VDtwY8T", "https://skylab-api.staging.amplitude.com", connTimeout), 
	// 	api,
	// 	flagConfigStorage, cohortStorage, nil)
	// println("inited")
	// // time.Sleep(5000 * time.Millisecond)
	// dr.start()

    // for {
        // fmt.Printf("%v+\n", time.Now())
		// fmt.Println(flagConfigStorage.GetFlagConfigs())
        // time.Sleep(5000 * time.Millisecond)
		// fmt.Println(flagConfigStorage.GetFlagConfigs())
    // }

	// if len(os.Args) < 2 {
	// 	fmt.Printf("error: command required\n")
	// 	fmt.Printf("Available commands:\n" +
	// 		"  fetch\n" +
	// 		"  flags\n" +
	// 		"  evaluate\n")
	// 	return
	// }
	// switch os.Args[1] {
	// case "fetch":
	// 	fetch()
	// case "flags":
	// 	flags()
	// case "evaluate":
	// 	evaluate()
	// default:
	// 	fmt.Printf("error: unknown sub-command '%v'", os.Args[1])
	// }
}