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

func runFlag(mgmtClient *managementClient) {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: xmpt flag <subcommand>\n")
		fmt.Printf("Subcommands: ls, create, get, edit, enable, disable, archive, unarchive, rollout\n")
		fmt.Printf("            list-user-ids, add-user-ids, delete-user-ids\n")
		fmt.Printf("            list-cohort-ids, add-cohort-ids, delete-cohort-ids\n")
		fmt.Printf("            list-deployments, add-deployments, delete-deployments\n")
		fmt.Printf("            list-deployment-ids, add-deployment-ids, delete-deployment-ids\n")
		os.Exit(1)
	}

	subcmd := os.Args[2]

	ctx := context.Background()

	switch subcmd {
	case "ls", "list":
		flagList(ctx, mgmtClient)
	case "create":
		flagCreate(ctx, mgmtClient)
	case "get":
		flagGet(ctx, mgmtClient)
	case "edit":
		flagEdit(ctx, mgmtClient)
	case "enable":
		flagEnable(ctx, mgmtClient)
	case "disable":
		flagDisable(ctx, mgmtClient)
	case "archive":
		flagArchive(ctx, mgmtClient)
	case "unarchive":
		flagUnarchive(ctx, mgmtClient)
	case "rollout":
		flagRollout(ctx, mgmtClient)
	case "list-user-ids":
		flagListUserIds(ctx, mgmtClient)
	case "add-user-ids":
		flagAddUserIds(ctx, mgmtClient)
	case "delete-user-ids":
		flagDeleteUserIds(ctx, mgmtClient)
	case "list-cohort-ids":
		flagListCohortIds(ctx, mgmtClient)
	case "add-cohort-ids":
		flagAddCohortIds(ctx, mgmtClient)
	case "delete-cohort-ids":
		flagDeleteCohortIds(ctx, mgmtClient)
	case "list-deployments":
		flagListDeployments(ctx, mgmtClient)
	case "add-deployments":
		flagAddDeployments(ctx, mgmtClient)
	case "delete-deployments":
		flagDeleteDeployments(ctx, mgmtClient)
	case "list-deployment-ids":
		flagListDeploymentIds(ctx, mgmtClient)
	case "add-deployment-ids":
		flagAddDeploymentIds(ctx, mgmtClient)
	case "delete-deployment-ids":
		flagDeleteDeploymentIds(ctx, mgmtClient)
	default:
		fmt.Printf("error: unknown flag subcommand '%v'\n", subcmd)
		os.Exit(1)
	}
}

// Parse a flag selector string into a project ID (optional) and flag key.
func parseFlagsWithFlagSelector(cmd *flag.FlagSet, args []string) ([]string, []string, []string, []string) {
	var projectIds arrayFlags
	cmd.Var(&projectIds, "project-id", "Project IDs (comma-separated)")
	var keys arrayFlags
	cmd.Var(&keys, "key", "Flag keys (comma-separated)")
	var ids arrayFlags
	cmd.Var(&ids, "id", "Flag IDs (comma-separated)")
	posArgs := parseFlags(cmd, args)

	if len(posArgs) == 0 || !keys.isEmpty() || !ids.isEmpty() || !strings.HasPrefix(posArgs[0], "@") {
		return projectIds.toList(), keys.toList(), ids.toList(), posArgs
	}

	flagSelector := posArgs[0][1:]

	parts := strings.Split(flagSelector, "/")
	if len(parts) >= 2 {
		// Check first part is a valid project id
		projectId := parts[0]
		if _, err := strconv.ParseInt(projectId, 10, 64); err != nil {
			// Not a project id, so it's a flag key as a whole
			keys.Set(flagSelector)
		} else {
			key := strings.Join(parts[1:], "/")
			projectIds.Set(parts[0])
			keys.Set(key)
		}
	} else {
		keys.Set(flagSelector)
	}
	return projectIds.toList(), keys.toList(), ids.toList(), posArgs[1:]
}

func flagList(ctx context.Context, client *managementClient) {
	listCmd := flag.NewFlagSet("flag list", flag.ExitOnError)
	var listFlags bool
	listCmd.BoolVar(&listFlags, "list", false, "List flag keys only")
	listCmd.BoolVar(&listFlags, "l", false, "List flag keys only")

	projectIds, flagKeys, flagIds, _ := parseFlagsWithFlagSelector(listCmd, os.Args[3:])
	if len(flagKeys) == 0 {
		// Allows to list all flags in all projects
		flagKeys = []string{""}
	}
	allFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(allFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	if listFlags {
		for _, flag := range allFlags {
			fmt.Printf("%s\t%s\t%s\t%s\n", *flag.Id, *flag.ProjectId, *flag.Key, *flag.Name)
		}
	} else {
		b, _ := json.Marshal(allFlags)
		fmt.Printf("%s\n", string(b))
	}
}

func flagCreate(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag create [OPTIONS] [RAW_JSON]\n")
		os.Exit(1)
	}

	createCmd := flag.NewFlagSet("flag create", flag.ExitOnError)
	flagKey := createCmd.String("key", "", "Flag key (required)")

	var projectIds arrayFlags
	createCmd.Var(&projectIds, "project-id", "Project IDs (comma-separated), if not specified, inferred from deploymentIds or deploymentNames")
	var deploymentNames arrayFlags
	createCmd.Var(&deploymentNames, "deployment", "Deployment names (comma-separated)")
	var deploymentIds arrayFlags
	createCmd.Var(&deploymentIds, "deployment-id", "Deployment IDs (comma-separated), overrides deployment names")

	name := createCmd.String("name", "", "Flag name")
	description := createCmd.String("description", "Created by xpmt", "Flag description")
	var variants arrayFlags
	createCmd.Var(&variants, "variant", "Variants (comma-separated or JSON)")
	bucketingKey := createCmd.String("bucketing-key", "", "Bucketing key (default to amplitude_id for remote evaluation and device_id for local evaluation)")
	var rolloutWeights arrayFlags
	createCmd.Var(&rolloutWeights, "rollout-weight", "Rollout weights (comma-separated variant=weight or JSON)")
	targetSegments := createCmd.String("target-segments", "", "Target segments JSON")
	evaluationMode := createCmd.String("evaluation-mode", "", "Evaluation mode (local|remote)")

	yes := createCmd.Bool("yes", false, "Yes to all")

	_ = parseFlags(createCmd, os.Args[3:])

	var rawJSON string
	if len(createCmd.Args()) > 0 {
		rawJSON = createCmd.Args()[0]
	}

	flagObj := &managementFlag{}

	if rawJSON != "" {
		err := json.Unmarshal([]byte(rawJSON), flagObj)
		if err != nil {
			fmt.Printf("error: invalid JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		if *flagKey == "" {
			fmt.Printf("error: --key is required\n")
			os.Exit(1)
		}
		flagObj.Key = flagKey
		if *name != "" {
			flagObj.Name = name
		} else {
			flagObj.Name = flagKey
		}
		flagObj.Description = description
		if !variants.isEmpty() {
			variantsList := variants.toList()
			variants := make([]interface{}, len(variantsList))
			for i, v := range variantsList {
				variants[i] = v
			}
			flagObj.Variants = &variants
		}
		if *bucketingKey == "" {
			if *evaluationMode == "remote" {
				key := "amplitude_id"
				bucketingKey = &key
			} else {
				key := "device_id"
				bucketingKey = &key
			}
		}
		flagObj.BucketingKey = bucketingKey
		if !rolloutWeights.isEmpty() {
			weights := make(map[string]int)
			for _, weight := range rolloutWeights.toList() {
				parts := strings.Split(weight, "=")
				weight, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("error: invalid weight for %s: %v\n", parts[0], err)
					os.Exit(1)
				}
				weights[parts[0]] = weight
			}
			flagObj.RolloutWeights = &weights
		}
		if *targetSegments != "" {
			var segments interface{}
			err := json.Unmarshal([]byte(*targetSegments), &segments)
			if err != nil {
				fmt.Printf("error: invalid targetSegments JSON: %v\n", err)
				os.Exit(1)
			}
			flagObj.TargetSegments = &segments
		}
		flagObj.EvaluationMode = evaluationMode

		// Parse project ids and deployments
		var parsedProjectIdsToDeployments = make(map[string]map[string]*managementDeployment, 0)
		if !projectIds.isEmpty() {
			for _, projectId := range projectIds.toList() {
				parsedProjectIdsToDeployments[projectId] = make(map[string]*managementDeployment, 0)
			}
			// ProjectId resolved. Now need to get deployments
			if !deploymentIds.isEmpty() {
				// Resolve deployment by id
				for _, depId := range deploymentIds.toList() {
					dep, err := client.getDeployment(ctx, depId)
					if err != nil {
						fmt.Printf("error: deployment id not found: %s: %v\n", depId, err)
						os.Exit(1)
					}
					if parsedProjectIdsToDeployments[*dep.ProjectId] == nil {
						// Don't add new project ids
						fmt.Printf("ignored deployment %s: project id not specified for deployment\n", depId)
						continue
					}
					parsedProjectIdsToDeployments[*dep.ProjectId][*dep.Id] = dep
				}
			} else if !deploymentNames.isEmpty() {
				// Resolve deployment by name, get deployments with label for all projects, then only match project ids specified
				for _, name := range deploymentNames.toList() {
					deps, err := client.getDeploymentByProjectIdLabel(ctx, nil, &name)
					if err != nil {
						fmt.Printf("error: deployment name not found: %s: %v\n", name, err)
						os.Exit(1)
					}
					for _, dep := range deps {
						if parsedProjectIdsToDeployments[*dep.ProjectId] == nil {
							// Don't add new project ids
							fmt.Printf("ignored deployment name %s for project id %s: project id not specified for deployment\n", name, *dep.ProjectId)
							continue
						}
						parsedProjectIdsToDeployments[*dep.ProjectId][*dep.Id] = dep
					}
				}
			}
		} else if !deploymentIds.isEmpty() {
			// Infer project IDs from deployment ids
			for _, depId := range deploymentIds.toList() {
				dep, err := client.getDeployment(ctx, depId)
				if err != nil {
					fmt.Printf("error: deployment id not found: %s: %v\n", depId, err)
					os.Exit(1)
				}
				if parsedProjectIdsToDeployments[*dep.ProjectId] == nil {
					// Add new project ids
					parsedProjectIdsToDeployments[*dep.ProjectId] = make(map[string]*managementDeployment, 0)
				}
				parsedProjectIdsToDeployments[*dep.ProjectId][*dep.Id] = dep
			}
		} else if !deploymentNames.isEmpty() {
			// Infer project IDs from deployment names
			for _, name := range deploymentNames.toList() {
				deps, err := client.getDeploymentByProjectIdLabel(ctx, nil, &name)
				if err != nil {
					fmt.Printf("error: deployment name not found: %s: %v\n", name, err)
					os.Exit(1)
				}
				for _, dep := range deps {
					if parsedProjectIdsToDeployments[*dep.ProjectId] == nil {
						// Add new project ids
						parsedProjectIdsToDeployments[*dep.ProjectId] = make(map[string]*managementDeployment, 0)
					}
					parsedProjectIdsToDeployments[*dep.ProjectId][*dep.Id] = dep
				}
			}
		} else {
			fmt.Printf("error: no project id can be inferred\n")
			os.Exit(1)
		}
		// Now we have a map of project ids to deployments
		// Looks like this:
		// {
		//   "projectId1": {
		//     "deploymentKey1": *managementDeployment,
		//   },
		// }
		// Create flags for each project id and its own subset of deployments

		flagsToCreate := make([]*managementFlag, 0, len(parsedProjectIdsToDeployments))
		for projectId, deployments := range parsedProjectIdsToDeployments {
			// Check flag key exists in any project
			flagKeyExists := false
			cursor := ""
			for {
				existingFlag, err := client.listFlags(ctx, flagObj.Key, &projectId, cursor)
				if err != nil {
					fmt.Printf("error: %v\n", err)
					os.Exit(1)
				}
				if len(existingFlag.Flags) > 0 {
					flagKeyExists = true
					break
				}
				cursor = existingFlag.NextCursor
				if cursor == "" {
					break
				}
			}
			if flagKeyExists {
				fmt.Printf("flag key %s already exists in project %s, skipping\n", *flagObj.Key, projectId)
				continue
			}

			// Construct flag object for creation
			newFlagObj := *flagObj
			newFlagObj.ProjectId = &projectId

			deploymentIds := make([]string, 0, len(deployments))
			for deploymentId := range deployments {
				deploymentIds = append(deploymentIds, deploymentId)
			}
			if len(deploymentIds) > 0 {
				newFlagObj.Deployments = &deploymentIds
			}
			flagsToCreate = append(flagsToCreate, &newFlagObj)

			b, _ := json.Marshal(newFlagObj)
			fmt.Printf("create flag %s\n", string(b))
		}

		if len(flagsToCreate) == 0 {
			fmt.Printf("no flags to create\n")
			return
		}

		confirm(*yes)

		for _, flag := range flagsToCreate {
			b, _ := json.Marshal(flag)
			fmt.Printf("creating flag %s\n", string(b))
			data, err := client.createFlag(ctx, flag)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("created flag %s in project %s\n", data.Id, *flag.ProjectId)
		}
	}
}

func flagGet(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag get FLAG_SELECTOR\n")
		os.Exit(1)
	}

	flagKey := os.Args[3]
	flagObj, err := client.getFlag(ctx, flagKey)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	b, _ := json.Marshal(flagObj)
	fmt.Printf("%s\n", string(b))
}

func filterFlags(ctx context.Context, client *managementClient, projectIds []string, flagKeys []string, flagIds []string) []*managementFlag {
	resultFlags := make([]*managementFlag, 0)
	if len(flagIds) == 0 {
		if len(projectIds) == 0 {
			projectIds = []string{""}
		}
		for _, projectId := range projectIds {
			for _, flagKey := range flagKeys {
				cursor := ""
				for {
					flags, err := client.listFlags(ctx, &flagKey, &projectId, cursor)
					if err != nil {
						fmt.Printf("error: %v\n", err)
						os.Exit(1)
					}
					for _, flag := range flags.Flags {
						resultFlags = append(resultFlags, &flag)
					}
					cursor = flags.NextCursor
					if cursor == "" {
						break
					}
				}
			}
		}
	} else {
		for _, flagId := range flagIds {
			flag, err := client.getFlag(ctx, flagId)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
			resultFlags = append(resultFlags, flag)
		}
	}

	return resultFlags
}

func confirm(yes bool) bool {
	if yes {
		fmt.Printf("--yes specified, performing operation\n")
		return true
	}
	fmt.Printf("Are you sure you want to perform the operation? (yes/no): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		fmt.Printf("operation cancelled\n")
		os.Exit(1)
	}
	return true
}

func flagEdit(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag edit FLAG_SELECTOR [OPTIONS] [RAW_JSON]\n")
		os.Exit(1)
	}

	editCmd := flag.NewFlagSet("flag edit", flag.ContinueOnError)
	name := editCmd.String("name", "", "Flag name")
	description := editCmd.String("description", "", "Flag description")
	bucketingKey := editCmd.String("bucketing-key", "", "Bucketing key")
	bucketingSalt := editCmd.String("bucketing-salt", "", "Bucketing salt")
	bucketingUnit := editCmd.String("bucketing-unit", "", "Bucketing unit")
	evaluationMode := editCmd.String("evaluation-mode", "", "Evaluation mode (local|remote)")
	rolloutPercentage := editCmd.Int("rollout-percentage", -1, "Rollout percentage")
	targetSegments := editCmd.String("target-segments", "", "Target segments JSON")
	enabled := editCmd.String("enabled", "", "Enable flag (true|false)")
	archive := editCmd.String("archive", "", "Archive flag (true|false)")
	var tags arrayFlags
	editCmd.Var(&tags, "tags", "Tags (comma-separated or JSON)")
	yes := editCmd.Bool("yes", false, "Yes to all")

	projectIds, flagKeys, flagIds, _ := parseFlagsWithFlagSelector(editCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	for _, flag := range foundFlags {
		fmt.Printf("update flag: %s\n", *flag.Id)
	}
	confirm(*yes)

	var rawJSON string
	if len(editCmd.Args()) > 0 {
		rawJSON = editCmd.Args()[0]
	}

	updates := &managementFlag{}

	if rawJSON != "" {
		err := json.Unmarshal([]byte(rawJSON), updates)
		if err != nil {
			fmt.Printf("error: invalid JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		if *name != "" {
			updates.Name = name
		}
		if *description != "" {
			updates.Description = description
		}
		if *bucketingKey != "" {
			updates.BucketingKey = bucketingKey
		}
		if *bucketingSalt != "" {
			updates.BucketingSalt = bucketingSalt
		}
		if *bucketingUnit != "" {
			updates.BucketingUnit = bucketingUnit
		}
		if *evaluationMode != "" {
			updates.EvaluationMode = evaluationMode
		}
		if *rolloutPercentage != -1 {
			updates.RolloutPercentage = rolloutPercentage
		}
		if *targetSegments != "" {
			var segments interface{}
			err := json.Unmarshal([]byte(*targetSegments), &segments)
			if err != nil {
				fmt.Printf("error: invalid targetSegments JSON: %v\n", err)
				os.Exit(1)
			}
			updates.TargetSegments = &segments
		}

		if *enabled != "" {
			enabled := *enabled == "true"
			updates.Enabled = &enabled
		}

		if *archive != "" {
			archive := *archive == "true"
			updates.Archive = &archive
		}

		if len(tags.toList()) > 0 {
			tagsList := tags.toList()
			updates.Tags = &tagsList
		}
	}

	for _, flag := range foundFlags {
		err := client.updateFlag(ctx, *flag.Id, updates)
		if err != nil {
			fmt.Printf("error while updating flag %s: %v\n", *flag.Id, err)
			os.Exit(1)
		}
		fmt.Printf("updated flag %s\n", *flag.Id)
	}
}

func flagEnable(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag enable FLAG_SELECTOR\n")
		os.Exit(1)
	}
	os.Args[2] = "edit"
	os.Args = append(os.Args, "--enabled", "true")

	flagEdit(ctx, client)
}

func flagDisable(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag disable FLAG_SELECTOR\n")
		os.Exit(1)
	}
	os.Args[2] = "edit"
	os.Args = append(os.Args, "--enabled", "false")

	flagEdit(ctx, client)
}

func flagArchive(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag archive FLAG_SELECTOR\n")
		os.Exit(1)
	}

	os.Args[2] = "edit"
	os.Args = append(os.Args, "--archive", "true")

	flagEdit(ctx, client)
}

func flagUnarchive(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag unarchive FLAG_SELECTOR\n")
		os.Exit(1)
	}

	os.Args[2] = "edit"
	os.Args = append(os.Args, "--archive", "false")

	flagEdit(ctx, client)
}

func flagRollout(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag rollout FLAG_SELECTOR ROLLOUT_PERCENTAGE\n")
		os.Exit(1)
	}
	origArgs := os.Args

	rolloutCmd := flag.NewFlagSet("flag rollout", flag.ContinueOnError)
	_, _, _, posArgs := parseFlagsWithFlagSelector(rolloutCmd, os.Args[3:])
	rolloutPercentage := posArgs[0]

	filteredArgs := make([]string, 0, len(origArgs))
	for _, arg := range origArgs {
		if arg == rolloutPercentage {
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	filteredArgs[2] = "edit"
	os.Args = append(filteredArgs, "--rollout-percentage", rolloutPercentage)

	flagEdit(ctx, client)
}

func parseFlagsWithVariantSelector(cmd *flag.FlagSet, args []string) ([]string, []string, []string, []string, []string) {
	variant := cmd.String("variant", "", "Variant key")
	projectIds, rawFlagKeys, flagIds, posArgs := parseFlagsWithFlagSelector(cmd, args)

	variantKeys := make([]string, 0)
	if *variant != "" {
		variantKeys = append(variantKeys, *variant)
		return projectIds, rawFlagKeys, flagIds, variantKeys, posArgs
	}

	// Process flagKeys as it may contain variant selectors
	flagKeys := make([]string, 0)
	for _, flagKey := range rawFlagKeys {
		parts := strings.SplitN(flagKey, "/", 2)
		if len(parts) >= 2 {
			variantKeys = append(variantKeys, strings.Join(parts[1:], "/"))
		}
		flagKeys = append(flagKeys, parts[0])
	}
	return projectIds, flagKeys, flagIds, variantKeys, posArgs
}

func flagListUserIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag list-user-ids VARIANT_SELECTOR\n")
		os.Exit(1)
	}

	variantAction(ctx, client, "list-user-ids")
}

func variantAction(ctx context.Context, client *managementClient, action string) {
	variantActionCmd := flag.NewFlagSet("flag variant action", flag.ContinueOnError)
	yes := variantActionCmd.Bool("yes", false, "Yes to all")
	projectIds, flagKeys, flagIds, variantKeys, posArgs := parseFlagsWithVariantSelector(variantActionCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}
	if len(variantKeys) == 0 {
		fmt.Printf("error: no variant keys found\n")
		os.Exit(1)
	}

	for _, flag := range foundFlags {
		for _, variantKey := range variantKeys {
			switch action {
			case "add-user-ids":
				fmt.Printf("add following user IDs to flag %s variant %s: %s\n", *flag.Id, variantKey, strings.Join(posArgs, ", "))
			case "delete-user-ids":
				fmt.Printf("delete following user IDs from flag %s variant %s: %s\n", *flag.Id, variantKey, strings.Join(posArgs, ", "))
			case "add-cohort-ids":
				fmt.Printf("add following cohort IDs to flag %s variant %s: %s\n", *flag.Id, variantKey, strings.Join(posArgs, ", "))
			case "delete-cohort-ids":
				fmt.Printf("delete following cohort IDs from flag %s variant %s: %s\n", *flag.Id, variantKey, strings.Join(posArgs, ", "))
			}
		}
	}
	switch action {
	case "add-user-ids":
		confirm(*yes)
	case "delete-user-ids":
		confirm(*yes)
	case "add-cohort-ids":
		confirm(*yes)
	case "delete-cohort-ids":
		confirm(*yes)
	}

	for _, flag := range foundFlags {
		for _, variantKey := range variantKeys {
			var err error
			switch action {
			case "list-user-ids":
				var userIds []string
				userIds, err = client.listVariantUserIds(ctx, *flag.Id, variantKey)
				if err == nil {
					fmt.Printf("user IDs for flag %s: %s\n", *flag.Id, strings.Join(userIds, ", "))
				}
			case "add-user-ids":
				err = client.addVariantUserIds(ctx, *flag.Id, variantKey, posArgs)
				if err == nil {
					fmt.Printf("added user IDs for flag %s variant %s\n", *flag.Id, variantKey)
				}
			case "delete-user-ids":
				err = client.deleteVariantUserIds(ctx, *flag.Id, variantKey, posArgs)
				if err == nil {
					fmt.Printf("deleted user IDs for flag %s variant %s\n", *flag.Id, variantKey)
				}
			case "list-cohort-ids":
				var cohortIds []string
				cohortIds, err = client.listVariantCohortIds(ctx, *flag.Id, variantKey)
				if err == nil {
					fmt.Printf("cohort IDs for flag %s: %s\n", *flag.Id, strings.Join(cohortIds, ", "))
				}
			case "add-cohort-ids":
				err = client.addVariantCohortIds(ctx, *flag.Id, variantKey, posArgs)
				if err == nil {
					fmt.Printf("added cohort IDs for flag %s variant %s\n", *flag.Id, variantKey)
				}
			case "delete-cohort-ids":
				err = client.deleteVariantCohortIds(ctx, *flag.Id, variantKey, posArgs)
				if err == nil {
					fmt.Printf("deleted cohort IDs for flag %s variant %s\n", *flag.Id, variantKey)
				}
			}
			if err != nil {
				fmt.Printf("error for flag %s variant %s: %v\n", *flag.Id, variantKey, err)
				os.Exit(1)
			}
		}
	}
}

func flagAddUserIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag add-user-ids VARIANT_SELECTOR user_id1 user_id2 ...\n")
		os.Exit(1)
	}

	variantAction(ctx, client, "add-user-ids")
}

func flagDeleteUserIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag delete-user-ids VARIANT_SELECTOR user_id1 user_id2 ...\n")
		os.Exit(1)
	}

	variantAction(ctx, client, "delete-user-ids")
}

func flagListCohortIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag list-cohort-ids VARIANT_SELECTOR\n")
		os.Exit(1)
	}

	variantAction(ctx, client, "list-cohort-ids")
}

func flagAddCohortIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag add-cohort-ids VARIANT_SELECTOR cohort1 cohort2 ...\n")
		os.Exit(1)
	}

	variantAction(ctx, client, "add-cohort-ids")
}

func flagDeleteCohortIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag delete-cohort-ids VARIANT_SELECTOR cohort1 cohort2 ...\n")
		os.Exit(1)
	}

	variantAction(ctx, client, "delete-cohort-ids")
}

func flagListDeployments(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag list-deployments FLAG_SELECTOR\n")
		os.Exit(1)
	}
	listDeploymentsCmd := flag.NewFlagSet("flag list-deployments", flag.ContinueOnError)

	projectIds, flagKeys, flagIds, _ := parseFlagsWithFlagSelector(listDeploymentsCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	deployments := make(map[string]interface{})
	for _, flag := range foundFlags {
		deploymentIds := flag.Deployments
		for _, id := range *deploymentIds {
			if _, exists := deployments[id]; !exists {
				deployment, err := client.getDeployment(ctx, id)
				if err != nil {
					fmt.Printf("error: failed to get deployment %s: %v\n", id, err)
					os.Exit(1)
				}
				deployments[id] = deployment
			}
		}
	}

	for _, flag := range foundFlags {
		deploymentLabelsForFlag := make([]string, 0, len(*flag.Deployments))
		for _, id := range *flag.Deployments {
			if deployment, exists := deployments[id]; exists {
				deploymentLabelsForFlag = append(deploymentLabelsForFlag, *deployment.(*managementDeployment).Label)
			}
		}
		fmt.Printf("deployments for flag %s in project %s: %s\n", *flag.Id, *flag.ProjectId, strings.Join(deploymentLabelsForFlag, ", "))
	}
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func flagAddDeployments(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag add-deployments FLAG_SELECTOR deployment1 deployment2 ...\n")
		os.Exit(1)
	}

	addDeploymentsCmd := flag.NewFlagSet("flag add-deployments", flag.ContinueOnError)
	yes := addDeploymentsCmd.Bool("yes", false, "Yes to all")
	projectIds, flagKeys, flagIds, posArgs := parseFlagsWithFlagSelector(addDeploymentsCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	newDeploymentIdsPerFlag := make(map[string][]string)
	for _, flag := range foundFlags {
		for _, label := range posArgs {
			deployment, err := client.getDeploymentByProjectIdLabel(ctx, flag.ProjectId, &label)
			if err != nil {
				fmt.Printf("error: deployment not found for flag %s in project %s: %s: %v\n", *flag.Id, *flag.ProjectId, label, err)
				continue
			}
			for _, d := range deployment {
				if containsString(*flag.Deployments, *d.Id) {
					continue
				}
				newDeploymentIdsPerFlag[*flag.Id] = append(newDeploymentIdsPerFlag[*flag.Id], *d.Id)
			}
			if len(newDeploymentIdsPerFlag[*flag.Id]) > 0 {
				fmt.Printf("add deployments %s to flag %s\n", strings.Join(newDeploymentIdsPerFlag[*flag.Id], ", "), *flag.Id)
			}
		}
	}
	if len(newDeploymentIdsPerFlag) == 0 {
		fmt.Printf("error: no new deployments found\n")
		os.Exit(1)
	}

	confirm(*yes)

	for flagId, deploymentIds := range newDeploymentIdsPerFlag {
		err := client.addFlagDeployments(ctx, flagId, deploymentIds)
		if err != nil {
			fmt.Printf("error: failed to add deployments %s to flag %s: %v\n", strings.Join(deploymentIds, ", "), flagId, err)
			os.Exit(1)
		}
		fmt.Printf("added deployments %s to flag %s\n", strings.Join(deploymentIds, ", "), flagId)
	}
}

func flagDeleteDeployments(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag delete-deployments FLAG_SELECTOR deployment1 deployment2 ...\n")
		os.Exit(1)
	}

	deleteDeploymentsCmd := flag.NewFlagSet("flag delete-deployments", flag.ContinueOnError)
	yes := deleteDeploymentsCmd.Bool("yes", false, "Yes to all")
	projectIds, flagKeys, flagIds, posArgs := parseFlagsWithFlagSelector(deleteDeploymentsCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	toDeleteDeploymentIdsPerFlag := make(map[string][]string)
	for _, flag := range foundFlags {
		for _, label := range posArgs {
			deployment, err := client.getDeploymentByProjectIdLabel(ctx, flag.ProjectId, &label)
			if err != nil {
				fmt.Printf("error: deployment not found for flag %s in project %s: %s: %v\n", *flag.Id, *flag.ProjectId, label, err)
				continue
			}
			for _, d := range deployment {
				if containsString(*flag.Deployments, *d.Id) {
					toDeleteDeploymentIdsPerFlag[*flag.Id] = append(toDeleteDeploymentIdsPerFlag[*flag.Id], *d.Id)
				}
			}
		}
		if len(toDeleteDeploymentIdsPerFlag[*flag.Id]) > 0 {
			fmt.Printf("delete deployments %s from flag %s\n", strings.Join(toDeleteDeploymentIdsPerFlag[*flag.Id], ", "), *flag.Id)
		}
	}
	if len(toDeleteDeploymentIdsPerFlag) == 0 {
		fmt.Printf("error: no deployments to delete found\n")
		os.Exit(1)
	}

	confirm(*yes)

	for flagId, deploymentIds := range toDeleteDeploymentIdsPerFlag {
		for _, deploymentId := range deploymentIds {
			err := client.deleteFlagDeployments(ctx, flagId, deploymentId)
			if err != nil {
				fmt.Printf("error: failed to delete deployment %s from flag %s: %v\n", deploymentId, flagId, err)
				os.Exit(1)
			}
			fmt.Printf("deleted deployment %s from flag %s\n", deploymentId, flagId)
		}
	}
}

func flagListDeploymentIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: xmpt flag list-deployment-ids FLAG_SELECTOR\n")
		os.Exit(1)
	}
	listDeploymentIdsCmd := flag.NewFlagSet("flag list-deployment-ids", flag.ContinueOnError)

	projectIds, flagKeys, flagIds, _ := parseFlagsWithFlagSelector(listDeploymentIdsCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no deployment IDs found\n")
		os.Exit(1)
	}

	for _, flag := range foundFlags {
		fmt.Printf("deployment IDs for flag %s: %s\n", *flag.Id, strings.Join(*flag.Deployments, ", "))
	}
}

func flagAddDeploymentIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag add-deployment-ids FLAG_SELECTOR deployment_id1 deployment_id2 ...\n")
		os.Exit(1)
	}

	addDeploymentsCmd := flag.NewFlagSet("flag add-deployments", flag.ContinueOnError)
	yes := addDeploymentsCmd.Bool("yes", false, "Yes to all")
	projectIds, flagKeys, flagIds, posArgs := parseFlagsWithFlagSelector(addDeploymentsCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	newDeploymentIdsPerFlag := make(map[string][]string)
	for _, deploymentId := range posArgs {

		deployment, err := client.getDeployment(ctx, deploymentId)
		if err != nil {
			fmt.Printf("error: deployment not found: %v\n", err)
			continue
		}

		for _, flag := range foundFlags {
			if *flag.ProjectId == *deployment.ProjectId {
				newDeploymentIdsPerFlag[*flag.Id] = append(newDeploymentIdsPerFlag[*flag.Id], deploymentId)
			}
		}
	}
	if len(newDeploymentIdsPerFlag) == 0 {
		fmt.Printf("error: no new deployments found\n")
		os.Exit(1)
	}

	for _, flag := range foundFlags {
		fmt.Printf("add deployments %s to flag %s\n", strings.Join(newDeploymentIdsPerFlag[*flag.Id], ", "), *flag.Id)
	}

	confirm(*yes)

	for flagId, deploymentIds := range newDeploymentIdsPerFlag {
		err := client.addFlagDeployments(ctx, flagId, deploymentIds)
		if err != nil {
			fmt.Printf("error: failed to add deployments %s to flag %s: %v\n", strings.Join(deploymentIds, ", "), flagId, err)
			os.Exit(1)
		}
		fmt.Printf("added deployments %s to flag %s\n", strings.Join(deploymentIds, ", "), flagId)
	}
}

func flagDeleteDeploymentIds(ctx context.Context, client *managementClient) {
	if len(os.Args) < 5 {
		fmt.Printf("Usage: xmpt flag delete-deployment-ids FLAG_SELECTOR deployment_id1 deployment_id2 ...\n")
		os.Exit(1)
	}

	deleteDeploymentsCmd := flag.NewFlagSet("flag delete-deployments", flag.ContinueOnError)
	yes := deleteDeploymentsCmd.Bool("yes", false, "Yes to all")
	projectIds, flagKeys, flagIds, posArgs := parseFlagsWithFlagSelector(deleteDeploymentsCmd, os.Args[3:])
	foundFlags := filterFlags(ctx, client, projectIds, flagKeys, flagIds)
	if len(foundFlags) == 0 {
		fmt.Printf("error: no flags found\n")
		os.Exit(1)
	}

	toDeleteDeploymentIdsPerFlag := make(map[string][]string)
	for _, flag := range foundFlags {
		for _, deploymentId := range posArgs {
			if containsString(*flag.Deployments, deploymentId) {
				toDeleteDeploymentIdsPerFlag[*flag.Id] = append(toDeleteDeploymentIdsPerFlag[*flag.Id], deploymentId)
			}
		}
		if len(toDeleteDeploymentIdsPerFlag[*flag.Id]) > 0 {
			fmt.Printf("delete deployments %s from flag %s\n", strings.Join(toDeleteDeploymentIdsPerFlag[*flag.Id], ", "), *flag.Id)
		}
	}
	if len(toDeleteDeploymentIdsPerFlag) == 0 {
		fmt.Printf("error: no deployments to delete found\n")
		os.Exit(1)
	}

	confirm(*yes)

	for flagId, deploymentIds := range toDeleteDeploymentIdsPerFlag {
		for _, deploymentId := range deploymentIds {
			err := client.deleteFlagDeployments(ctx, flagId, deploymentId)
			if err != nil {
				fmt.Printf("error: failed to delete deployment %s from flag %s: %v\n", deploymentId, flagId, err)
				os.Exit(1)
			}
			fmt.Printf("deleted deployment %s from flag %s\n", deploymentId, flagId)
		}
	}
}
