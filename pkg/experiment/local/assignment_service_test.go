package local

import (
	"fmt"
	"github.com/amplitude/experiment-go-server/pkg/experiment"
	"testing"
	"time"
)

func TestToEvent(t *testing.T) {
	user := &experiment.User{
		UserId:   "user",
		DeviceId: "device",
	}

	results := &evaluationResult{
		"flag-key-1": flagResult{
			Variant: evaluationVariant{
				Key: "on",
			},
			IsDefaultVariant: false,
		},
		"flag-key-2": flagResult{
			Variant: evaluationVariant{
				Key: "control",
			},
			IsDefaultVariant: true,
		},
	}

	assignment := NewAssignment(user, results)
	event := toEvent(assignment)
	canonicalization := "user device flag-key-1 on flag-key-2 control "
	expectedInsertID := fmt.Sprintf("user device %d %d", hashCode(canonicalization), assignment.timestamp/DayMillis)
	if event.UserID != "user" {
		t.Errorf("UserID was %s, expected %s", event.UserID, "user")
	}
	if event.DeviceID != "device" {
		t.Errorf("DeviceID was %s, expected %s", event.DeviceID, "device")
	}
	if len(event.UserProperties) != 2 {
		t.Errorf("Length of UserProperties was %d, expected %d", len(event.UserProperties), 2)
	}
	if len(event.EventProperties) != 2 {
		t.Errorf("Length of EventProperties was %d, expected %d", len(event.EventProperties), 2)
	}
	if event.InsertID != expectedInsertID {
		t.Errorf("InsertID was %s, expected %s", event.InsertID, expectedInsertID)
	}

	fmt.Println(event.UserID)
	fmt.Println(event.DeviceID)
	fmt.Println(event.UserProperties)
	fmt.Println(event.EventProperties)
	fmt.Println(event.InsertID)
}

func TestCallLocalEvalWithAssignmentConfig(t *testing.T) {
	analyticsApiKey := "a6dd847b9d2f03c816d4f3f8458cdc1d"
	deploymentApiKey := "server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz"
	assignmentConfig := AssignmentConfig{ApiKey: analyticsApiKey}
	client := Initialize(deploymentApiKey, &Config{Debug: true, AssignmentConfig: assignmentConfig})
	client.Start()
	client.Evaluate(&experiment.User{UserId: "tim.yiu@amplitude.com"}, nil)
	(*client.assignmentService.Amplitude).Flush()
	time.Sleep(4 * time.Second)
}

func TestCallLocalEvalWithoutAssignmentConfig(t *testing.T) {
	deploymentApiKey := "server-qz35UwzJ5akieoAdIgzM4m9MIiOLXLoz"
	client := Initialize(deploymentApiKey, nil)
	client.Start()
	client.Evaluate(&experiment.User{UserId: "tim.yiu@amplitude.com"}, nil)
	time.Sleep(4 * time.Second)
}
