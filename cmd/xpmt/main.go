package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}
	mgmtKey := os.Getenv("MANAGEMENT_API_KEY")
	if mgmtKey == "" {
		fmt.Printf("error: MANAGEMENT_API_KEY environment variable required\n")
		os.Exit(1)
	}
	mgmtClient := newManagementClient(mgmtKey, nil)

	cmd := os.Args[1]
	switch cmd {
	case "eval", "evaluate":
		runEval(mgmtClient)
	case "flag", "flags":
		runFlag(mgmtClient)
	case "exp", "experiment":
		runExp()
	case "depl", "deployment":
		runDepl(mgmtClient)
	case "fetch":
		fetch()
	default:
		fmt.Printf("error: unknown command '%v'\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("xmpt COMMAND\n\n")
	fmt.Printf("  eval\n")
	fmt.Printf("    Evaluate flags for a user\n\n")
	fmt.Printf("  flag\n")
	fmt.Printf("    Manage flags\n\n")
	fmt.Printf("  exp\n")
	fmt.Printf("    Manage experiments\n\n")
	fmt.Printf("  depl\n")
	fmt.Printf("    Manage deployments\n")
}

func runExp() {
	fmt.Printf("Experiment management not yet implemented\n")
	os.Exit(1)
}
