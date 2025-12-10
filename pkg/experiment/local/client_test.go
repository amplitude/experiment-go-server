package local

import (
	"log"
	"os"
	"testing"

	"github.com/amplitude/analytics-go/amplitude"
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

func TestEvaluateV2WithTracksExposureTracksNonDefaultVariants(t *testing.T) {
	exposureConfig := &ExposureConfig{Config: amplitude.Config{APIKey: "some_api_key"}}
	clients["server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz"] = nil
	client = Initialize("server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz",
		&Config{ExposureConfig: exposureConfig})
	err := client.Start()
	if err != nil {
		panic(err)
	}

	user := &experiment.User{UserId: "test_user", DeviceId: "device_id"}

	// Capture tracked events
	trackedEvents := make([]amplitude.Event, 0)

	// Create a mock amplitude client that captures events
	mockAmplitudeClient := mockAmplitudeClientForTest{
		trackedEvents: &trackedEvents,
	}

	oldAmplitude := client.exposureService.amplitude
	client.exposureService.amplitude = &mockAmplitudeClient
	defer func() {
		client.exposureService.amplitude = oldAmplitude
	}()

	// Perform evaluation with TracksExposure=true
	options := EvaluateOptions{
		TracksExposure: true,
	}
	variants, err := client.EvaluateV2WithOptions(user, &options)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	// Verify that track was called
	if len(trackedEvents) == 0 {
		t.Fatalf("Expected exposure events to be tracked, but none were tracked")
	}

	// Count non-default variants
	nonDefaultVariants := make(map[string]experiment.Variant)
	for flagKey, variant := range variants {
		isDefault, ok := variant.Metadata["default"].(bool)
		if !ok || !isDefault {
			nonDefaultVariants[flagKey] = variant
		}
	}

	// Verify that we have one event per non-default variant
	if len(trackedEvents) != len(nonDefaultVariants) {
		t.Fatalf("Expected %d exposure events, got %d", len(nonDefaultVariants), len(trackedEvents))
	}

	// Verify each event has the correct structure
	trackedFlagKeys := make(map[string]bool)
	for _, event := range trackedEvents {
		if event.EventType != "[Experiment] Exposure" {
			t.Errorf("EventType was %s, expected %s", event.EventType, "[Experiment] Exposure")
		}
		if event.UserID != user.UserId {
			t.Errorf("UserID was %s, expected %s", event.UserID, user.UserId)
		}
		flagKey, ok := event.EventProperties["[Experiment] Flag Key"].(string)
		if !ok || flagKey == "" {
			t.Errorf("Event should have flag key")
		}
		trackedFlagKeys[flagKey] = true
		// Verify the variant is not default
		variant, exists := variants[flagKey]
		if !exists {
			t.Errorf("Variant for %s should exist", flagKey)
		}
		isDefault, ok := variant.Metadata["default"].(bool)
		if ok && isDefault {
			t.Errorf("Variant for %s should not be default", flagKey)
		}
	}

	// Verify all non-default variants were tracked
	if len(trackedFlagKeys) != len(nonDefaultVariants) {
		t.Errorf("Expected %d tracked flag keys, got %d", len(nonDefaultVariants), len(trackedFlagKeys))
	}
	for flagKey := range nonDefaultVariants {
		if !trackedFlagKeys[flagKey] {
			t.Errorf("Flag key %s should have been tracked", flagKey)
		}
	}
}

type mockAmplitudeClientForTest struct {
	trackedEvents *[]amplitude.Event
}

func (m *mockAmplitudeClientForTest) Track(event amplitude.Event) {
	*m.trackedEvents = append(*m.trackedEvents, event)
}
