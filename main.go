package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Assertions struct {
	Data []struct {
		StoreAssertionRecords []struct {
			AssertionDetails struct {
				ModeIdentifier string `json:"assertionDetailsModeIdentifier"`
			} `json:"assertionDetails"`
			AssertionStartDate float64 `json:"assertionStartDateTimestamp"`
		} `json:"storeAssertionRecords"`
	} `json:"data"`
}

type ModeConfigurations struct {
	Data []struct {
		ModeConfigurations map[string]struct {
			Mode struct {
				Name           string `json:"name"`
				ModeIdentifier string `json:"modeIdentifier"`
			} `json:"mode"`
		} `json:"modeConfigurations"`
	} `json:"data"`
}

func main() {
	outputFile := flag.String("output", "current_focus.txt", "Output file path")
	assertionsPath := flag.String("assertions", filepath.Join(os.Getenv("HOME"), "Library/DoNotDisturb/DB/Assertions.json"), "Path to Assertions.json")
	modesPath := flag.String("modes", filepath.Join(os.Getenv("HOME"), "Library/DoNotDisturb/DB/ModeConfigurations.json"), "Path to ModeConfigurations.json")
	flag.Parse()

	// Read and parse Assertions.json
	assertionsData, err := os.ReadFile(*assertionsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading assertions file: %v\n", err)
		os.Exit(1)
	}

	var assertions Assertions
	if err := json.Unmarshal(assertionsData, &assertions); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing assertions JSON: %v\n", err)
		os.Exit(1)
	}

	// Read and parse ModeConfigurations.json
	modesData, err := os.ReadFile(*modesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading mode configurations file: %v\n", err)
		os.Exit(1)
	}

	var modeConfigs ModeConfigurations
	if err := json.Unmarshal(modesData, &modeConfigs); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing mode configurations JSON: %v\n", err)
		os.Exit(1)
	}

	// Build a map of mode identifiers to names
	modeNames := make(map[string]string)
	if len(modeConfigs.Data) > 0 {
		for modeID, modeConfig := range modeConfigs.Data[0].ModeConfigurations {
			modeNames[modeID] = modeConfig.Mode.Name
		}
	}

	// Get the currently active assertion (from storeAssertionRecords)
	currentFocus := "None"
	if len(assertions.Data) > 0 && len(assertions.Data[0].StoreAssertionRecords) > 0 {
		// Get the most recent assertion (highest timestamp)
		var latestRecord *struct {
			AssertionDetails struct {
				ModeIdentifier string `json:"assertionDetailsModeIdentifier"`
			} `json:"assertionDetails"`
			AssertionStartDate float64 `json:"assertionStartDateTimestamp"`
		}

		for i := range assertions.Data[0].StoreAssertionRecords {
			record := &assertions.Data[0].StoreAssertionRecords[i]
			if latestRecord == nil || record.AssertionStartDate > latestRecord.AssertionStartDate {
				latestRecord = record
			}
		}

		if latestRecord != nil {
			modeID := latestRecord.AssertionDetails.ModeIdentifier
			if name, exists := modeNames[modeID]; exists {
				currentFocus = name
			} else {
				currentFocus = modeID
			}
		}
	}

	// Write to output file
	if err := os.WriteFile(*outputFile, []byte(currentFocus+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
}
