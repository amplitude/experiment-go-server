package local

import (
	"testing"
	"time"
)

// func TestFlagConfigStreamApi(t *testing.T) {
// 	api := NewFlagConfigStreamApiV2("deploymentkey", "serverurl", 1 * time.Second)
// 	api.newSseStreamFactory = func(authToken, url string, connectionTimeout, keepaliveTimeout, reconnInterval, maxJitter time.Duration, newESFactory NewEventSourceFactory) *SseStream {}
// 	api.Connect(
// 		func(m map[string]*evaluation.Flag) error {return nil},
// 		func(m map[string]*evaluation.Flag) error {return nil},
// 		func(err error) {},
// 	)

// }


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