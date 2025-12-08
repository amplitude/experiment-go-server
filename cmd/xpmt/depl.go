package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runDepl(mgmtClient *managementClient) {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: xmpt depl <subcommand>\n")
		fmt.Printf("Subcommands: list, create, delete\n")
		os.Exit(1)
	}

	subcmd := os.Args[2]
	ctx := context.Background()

	switch subcmd {
	case "list", "ls":
		deplList(ctx, mgmtClient)
	case "create":
		deplCreate(ctx, mgmtClient)
	case "edit":
		editDeployment(ctx, mgmtClient)
	case "archive":
		archiveDeployment(ctx, mgmtClient)
	case "unarchive":
		unarchiveDeployment(ctx, mgmtClient)
	default:
		fmt.Printf("error: unknown deployment subcommand '%v'\n", subcmd)
		os.Exit(1)
	}
}

func deplList(ctx context.Context, client *managementClient) {
	listCmd := flag.NewFlagSet("depl list", flag.ExitOnError)
	var list bool
	listCmd.BoolVar(&list, "list", false, "List deployment keys only")
	listCmd.BoolVar(&list, "l", false, "List deployment keys only")

	_ = parseFlags(listCmd, os.Args[3:])

	response, err := client.listDeployments(ctx, "")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	allDeployments := response.Deployments
	cursor := response.NextCursor
	for cursor != "" {
		response, err = client.listDeployments(ctx, cursor)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		allDeployments = append(allDeployments, response.Deployments...)
		cursor = response.NextCursor
	}

	if list {
		for _, deployment := range allDeployments {
			status := ""
			if deployment.Deleted != nil && *deployment.Deleted {
				status = "deleted"
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", *deployment.Id, *deployment.ProjectId, *deployment.Key, *deployment.Label, status)
		}
	} else {
		b, _ := json.Marshal(allDeployments)
		fmt.Printf("%s\n", string(b))
	}
}

func deplCreate(ctx context.Context, client *managementClient) {
	createCmd := flag.NewFlagSet("depl create", flag.ExitOnError)
	projectId := createCmd.String("project-id", "", "Project ID (required)")
	label := createCmd.String("label", "", "Deployment label (required)")
	deplType := createCmd.String("type", "client", "Deployment type (client, server) (required)")
	yes := createCmd.Bool("yes", false, "Yes to all")
	_ = parseFlags(createCmd, os.Args[3:])

	deployment := &managementDeployment{
		ProjectId: projectId,
		Label:     label,
		Type:      deplType,
	}

	fmt.Printf("create deployment with label %s and type %s in project %s\n", *deployment.Label, *deployment.Type, *deployment.ProjectId)
	confirm(*yes)

	created, err := client.createDeployment(ctx, deployment)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	b, _ := json.Marshal(created)
	fmt.Printf("%s\n", string(b))
}

// Parse a flag selector string into a project ID (optional) and flag key.
func parseDeploymentSelector(cmd *flag.FlagSet, args []string) ([]string, []string, []string, []string) {
	var projectIds arrayFlags
	cmd.Var(&projectIds, "project-id", "Project IDs (comma-separated)")
	var labels arrayFlags
	cmd.Var(&labels, "label", "Deployment labels (comma-separated)")
	var ids arrayFlags
	cmd.Var(&ids, "id", "Deployment IDs (comma-separated)")
	posArgs := parseFlags(cmd, args)

	if len(posArgs) == 0 || !labels.isEmpty() || !ids.isEmpty() || !strings.HasPrefix(posArgs[0], "@") {
		return projectIds.toList(), labels.toList(), ids.toList(), posArgs
	}

	deploymentSelector := posArgs[0][1:]

	parts := strings.Split(deploymentSelector, "/")
	if len(parts) >= 2 {
		// Check first part is a valid project id
		projectId := parts[0]
		if _, err := strconv.ParseInt(projectId, 10, 64); err != nil {
			// Not a project id, so it's a flag key as a whole
			labels.Set(deploymentSelector)
		} else {
			label := strings.Join(parts[1:], "/")
			projectIds.Set(parts[0])
			labels.Set(label)
		}
	} else {
		labels.Set(deploymentSelector)
	}
	return projectIds.toList(), labels.toList(), ids.toList(), posArgs[1:]
}

func filterDeployments(ctx context.Context, client *managementClient, projectIds []string, labels []string, deploymentIds []string) []*managementDeployment {
	allDeployments := make([]*managementDeployment, 0)
	cursor := ""
	for {
		response, err := client.listDeployments(ctx, cursor)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		allDeployments = append(allDeployments, response.Deployments...)
		cursor = response.NextCursor
		if cursor == "" {
			break
		}
	}

	foundDeployments := make([]*managementDeployment, 0)
	for _, deployment := range allDeployments {
		if containsString(deploymentIds, *deployment.Id) || ((len(projectIds) == 0 || containsString(projectIds, *deployment.ProjectId)) && (len(labels) == 0 || containsString(labels, *deployment.Label))) {
			foundDeployments = append(foundDeployments, deployment)
		}
	}

	return foundDeployments
}

func editDeployment(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt depl edit DEPLOYMENT_ID\n")
		os.Exit(1)
	}

	editCmd := flag.NewFlagSet("depl edit", flag.ExitOnError)
	// label := editCmd.String("label", "", "Deployment label")
	archive := editCmd.String("archive", "", "Archive deployment")
	yes := editCmd.Bool("yes", false, "Yes to all")

	projectIds, labels, deploymentIds, _ := parseDeploymentSelector(editCmd, os.Args[3:])
	foundDeployments := filterDeployments(ctx, client, projectIds, labels, deploymentIds)
	if len(foundDeployments) == 0 {
		fmt.Printf("error: no deployments found\n")
		os.Exit(1)
	}

	for _, deployment := range foundDeployments {
		fmt.Printf("edit deployment %s\n", *deployment.Id)
	}

	confirm(*yes)

	updates := &managementDeployment{}
	// if *label != "" {
	// 	updates.Label = label
	// }
	if *archive != "" {
		doArchive := *archive == "true"
		updates.Archive = &doArchive
	}

	for _, deployment := range foundDeployments {
		err := client.updateDeployment(ctx, *deployment.Id, updates)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("updated deployment %s\n", *deployment.Id)
	}
}

func archiveDeployment(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt depl archive DEPLOYMENT_ID\n")
		os.Exit(1)
	}

	os.Args[2] = "edit"
	os.Args = append(os.Args, "--archive", "true")
	editDeployment(ctx, client)
}

func unarchiveDeployment(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt depl unarchive DEPLOYMENT_ID\n")
		os.Exit(1)
	}

	os.Args[2] = "edit"
	os.Args = append(os.Args, "--archive", "false")
	editDeployment(ctx, client)
}
