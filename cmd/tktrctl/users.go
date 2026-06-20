package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users (admin only)",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		var users []map[string]any
		if err := get("users", &users); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(users, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var usersShowCmd = &cobra.Command{
	Use:   "show <user-id>",
	Short: "Show a single user by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var user map[string]any
		if err := get("users/"+args[0], &user); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(user, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var usersCreateCmd = &cobra.Command{
	Use:   "create <username>",
	Short: "Create a new user (admin only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]
		displayName, _ := cmd.Flags().GetString("display-name")
		isAdmin, _ := cmd.Flags().GetBool("admin")

		body := map[string]any{
			"username": username,
			"admin":    isAdmin,
		}
		if displayName != "" {
			body["display_name"] = displayName
		}

		var user map[string]any
		if err := post("users", body, &user); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(user, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var usersUpdateCmd = &cobra.Command{
	Use:   "update <user-id>",
	Short: "Update a user's display name or PAT (admin only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if _, err := strconv.Atoi(id); err != nil {
			return fmt.Errorf("user-id must be a numeric ID")
		}
		displayName, _ := cmd.Flags().GetString("display-name")
		userPat, _ := cmd.Flags().GetString("pat")

		if displayName == "" && userPat == "" {
			return fmt.Errorf("at least one of --display-name or --pat is required")
		}

		body := map[string]string{}
		if displayName != "" {
			body["display_name"] = displayName
		}
		if userPat != "" {
			body["pat"] = userPat
		}

		var user map[string]any
		if err := api("PATCH", "users/"+id, body, &user); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(user, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var usersDeleteCmd = &cobra.Command{
	Use:   "delete <user-id>",
	Short: "Delete a user (admin only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if _, err := strconv.Atoi(id); err != nil {
			return fmt.Errorf("user-id must be a numeric ID")
		}
		if err := api("DELETE", "users/"+id, nil, nil); err != nil {
			return err
		}
		fmt.Printf("user %s deleted\n", id)
		return nil
	},
}

func init() {
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersShowCmd)

	usersCreateCmd.Flags().StringP("display-name", "d", "", "Display name (defaults to username)")
	usersCreateCmd.Flags().Bool("admin", false, "Make the user an admin")
	usersCmd.AddCommand(usersCreateCmd)

	usersUpdateCmd.Flags().StringP("display-name", "d", "", "New display name")
	usersUpdateCmd.Flags().StringP("pat", "p", "", "New personal access token")
	usersCmd.AddCommand(usersUpdateCmd)

	usersCmd.AddCommand(usersDeleteCmd)
}
