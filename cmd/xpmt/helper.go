package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type arrayFlags struct {
	values []string
}

func newArrayFlags() *arrayFlags {
	return &arrayFlags{
		values: make([]string, 0),
	}
}

func (i *arrayFlags) toMap() map[string]string {
	props := make(map[string]string)

	var propString []string
	if len(i.values) == 1 {
		if i.values[0] == "{" {
			var v map[string]string
			err := json.Unmarshal([]byte(i.values[0]), &v)
			if err != nil {
				fmt.Printf("error: invalid JSON: %v\n", err)
				os.Exit(1)
			}
			return v
		}
		propString = strings.Split(i.values[0], ",")
	} else {
		propString = i.values
	}
	for _, prop := range propString {
		parts := strings.SplitN(prop, "=", 2)
		if len(parts) == 2 {
			props[parts[0]] = parts[1]
		} else {
			props[prop] = ""
		}
	}
	return props
}

func (i *arrayFlags) toList() []string {
	if len(i.values) == 1 {
		if i.values[0] == "[" {
			var v []string
			err := json.Unmarshal([]byte(i.values[0]), &v)
			if err != nil {
				fmt.Printf("error: invalid JSON: %v\n", err)
				os.Exit(1)
			}
			return v
		}
		return strings.Split(i.values[0], ",")
	}
	return i.values
}

// String is an implementation of the flag.Value interface
func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", i.values)
}

// Set is an implementation of the flag.Value interface
func (i *arrayFlags) Set(value string) error {
	i.values = append(i.values, value)
	return nil
}

func (i *arrayFlags) isEmpty() bool {
	return len(i.values) == 0
}

func parseFlags(cmd *flag.FlagSet, rawArgs []string) []string {
	argReplacementsJson := os.Getenv("XPMT_ARG_REPLACEMENTS")
	if argReplacementsJson == "" {
		argReplacementsJson = "{}"
	}
	var argReplacements map[string][]string
	err := json.Unmarshal([]byte(argReplacementsJson), &argReplacements)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	args := make([]string, 0, len(rawArgs))
	for _, arg := range rawArgs {
		if replacements, ok := argReplacements[arg]; ok {
			args = append(args, replacements...)
		} else {
			args = append(args, arg)
		}
	}

	posArgs := make([]string, 0, len(args))
	for {
		err := cmd.Parse(args)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			// os.Exit(1)
		}
		args = cmd.Args()
		if len(args) == 0 {
			break
		}
		posArgs = append(posArgs, args[0])
		args = args[1:]
	}
	return posArgs
}
