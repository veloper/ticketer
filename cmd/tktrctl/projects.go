package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		var projects []map[string]any
		if err := get("projects", &projects); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(projects, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <project-id>",
	Short: "Update a project's name, slug, or description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := resolveProject(args[0])
		if id == "" {
			return fmt.Errorf("project not found: %s", args[0])
		}
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		desc, _ := cmd.Flags().GetString("description")

		body := map[string]string{}
		if name != "" {
			body["name"] = name
		}
		if slug != "" {
			body["slug"] = slug
		}
		if desc != "" {
			body["description"] = desc
		}
		if len(body) == 0 {
			return fmt.Errorf("at least one of --name, --slug, or --description is required")
		}

		var project map[string]any
		if err := api("PATCH", "projects/"+id, body, &project); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(project, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var projectsShowCmd = &cobra.Command{
	Use:   "show <project-id>",
	Short: "Show a single project by ID or slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := resolveProject(args[0])
		if id == "" {
			return fmt.Errorf("project not found: %s", args[0])
		}
		var project map[string]any
		if err := get("projects/"+id, &project); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(project, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create <name> <slug>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		slug := args[1]
		description, _ := cmd.Flags().GetString("description")

		body := map[string]string{
			"name":        name,
			"slug":        slug,
			"description": description,
		}
		var project map[string]any
		if err := post("projects", body, &project); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(project, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsShowCmd)
	projectsCreateCmd.Flags().StringP("description", "d", "", "Project description")
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsUpdateCmd.Flags().StringP("name", "n", "", "New name")
	projectsUpdateCmd.Flags().StringP("slug", "s", "", "New slug")
	projectsUpdateCmd.Flags().StringP("description", "d", "", "New description")
	projectsCmd.AddCommand(projectsUpdateCmd)
}
