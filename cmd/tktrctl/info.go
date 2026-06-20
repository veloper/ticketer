package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show server info, users, and projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		var info map[string]any
		if err := get("info", &info); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}
