package local

import (
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
)

var clientEU *Client

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	projectApiKey := os.Getenv("EU_API_KEY")
	secretKey := os.Getenv("EU_SECRET_KEY")
	cohortSyncConfig := CohortSyncConfig{
		ApiKey:    projectApiKey,
		SecretKey: secretKey,
	}
	clientEU = Initialize("server-Qlp7XiSu6JtP2S3JzA95PnP27duZgQCF",
		&Config{CohortSyncConfig: &cohortSyncConfig, ServerZone: "eu"})
	err = clientEU.Start()
	if err != nil {
		panic(err)
	}
}

func TestEvaluateV2CohortEU(t *testing.T) {
	user := &experiment.User{UserId: "1", DeviceId: "0"}
	flagKeys := []string{"sdk-local-evaluation-user-cohort"}
	result, err := clientEU.EvaluateV2(user, flagKeys)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	variant := result["sdk-local-evaluation-user-cohort"]
	if variant.Key != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
	if variant.Value != "on" {
		t.Fatalf("Unexpected variant %v", variant)
	}
}
