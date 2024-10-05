package local

import (
	"log"
	"os"
	"testing"

	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var streamClient *Client

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
	streamClient = Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz",
		&Config{
			StreamUpdates: true,
			StreamServerUrl: "https://stream.lab.amplitude.com",
			CohortSyncConfig: &cohortSyncConfig,
		})
	err = streamClient.Start()
	if err != nil {
		panic(err)
	}
}

func TestMakeSureStreamEnabled(t *testing.T) {
	assert.True(t, streamClient.config.StreamUpdates)
}

func TestStreamEvaluate(t *testing.T) {
	user := &experiment.User{UserId: "test_user"}
	result, err := streamClient.Evaluate(user, nil)
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

func TestStreamEvaluateV2AllFlags(t *testing.T) {
	user := &experiment.User{UserId: "test_user"}
	result, err := streamClient.EvaluateV2(user, nil)
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

func TestStreamFlagMetadataLocalFlagKey(t *testing.T) {
	md := streamClient.FlagMetadata("sdk-local-evaluation-ci-test")
	if md["evaluationMode"] != "local" {
		t.Fatalf("Unexpected metadata %v", md)
	}
}

func TestStreamEvaluateV2Cohort(t *testing.T) {
	targetedUser := &experiment.User{UserId: "12345"}
	nonTargetedUser := &experiment.User{UserId: "not_targeted"}
	flagKeys := []string{"sdk-local-evaluation-user-cohort-ci-test"}
	result, err := streamClient.EvaluateV2(targetedUser, flagKeys)
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
	result, err = streamClient.EvaluateV2(nonTargetedUser, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant = result["sdk-local-evaluation-user-cohort-ci-test"]
	if variant.Key != "off" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}

func TestStreamEvaluateV2GroupCohort(t *testing.T) {
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
	result, err := streamClient.EvaluateV2(targetedUser, flagKeys)
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
	result, err = streamClient.EvaluateV2(nonTargetedUser, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant = result["sdk-local-evaluation-group-cohort-ci-test"]
	if variant.Key != "off" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}
