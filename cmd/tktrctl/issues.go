package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var issuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "Manage issues",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var issuesListCmd = &cobra.Command{
	Use:   "list <project-id>",
	Short: "List issues in a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]
		// Resolve project slug to ID if needed
		id := resolveProject(projectID)
		if id == "" {
			return fmt.Errorf("project not found: %s", projectID)
		}

		state, _ := cmd.Flags().GetString("state")
		assignee, _ := cmd.Flags().GetString("assignee")
		path := "projects/" + id + "/issues"
		if state != "" || assignee != "" {
			path += "?"
			if state != "" {
				path += "state=" + state
			}
			if assignee != "" {
				if state != "" {
					path += "&"
				}
				path += "assignee=" + assignee
			}
		}

		var issues []map[string]any
		if err := get(path, &issues); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(issues, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var issuesShowCmd = &cobra.Command{
	Use:   "show <issue-id>",
	Short: "Show a single issue by ID or slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var issue map[string]any
		if err := get("issues/"+args[0], &issue); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(issue, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var issuesCreateCmd = &cobra.Command{
	Use:   "create <project-id> <title>",
	Short: "Create a new issue in a project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]
		title := args[1]
		id := resolveProject(projectID)
		if id == "" {
			return fmt.Errorf("project not found: %s", projectID)
		}
		description, _ := cmd.Flags().GetString("description")
		typ, _ := cmd.Flags().GetString("type")
		state, _ := cmd.Flags().GetString("state")
		priority, _ := cmd.Flags().GetInt("priority")

		body := map[string]any{
			"title":       title,
			"description": description,
			"type":        typ,
			"state":       state,
			"priority":    priority,
		}
		if assignee, _ := cmd.Flags().GetInt64("assignee"); assignee != 0 {
			body["assignee"] = assignee
		}

		var issue map[string]any
		if err := post("projects/"+id+"/issues", body, &issue); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(issue, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var issuesUpdateCmd = &cobra.Command{
	Use:   "update <issue-id>",
	Short: "Update an issue's fields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		title, _ := cmd.Flags().GetString("title")
		desc, _ := cmd.Flags().GetString("description")
		typ, _ := cmd.Flags().GetString("type")
		state, _ := cmd.Flags().GetString("state")
		priority, _ := cmd.Flags().GetInt("priority")
		assignee, _ := cmd.Flags().GetInt64("assignee")

		body := map[string]any{}
		if title != "" {
			body["title"] = title
		}
		if desc != "" {
			body["description"] = desc
		}
		if typ != "" {
			body["type"] = typ
		}
		if state != "" {
			body["state"] = state
		}
		if priority != 0 {
			body["priority"] = priority
		}
		if assignee != 0 {
			body["assignee"] = assignee
		}
		if len(body) == 0 {
			return fmt.Errorf("at least one flag is required")
		}

		var issue map[string]any
		if err := api("PATCH", "issues/"+issueID, body, &issue); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(issue, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

var issuesStateCmd = &cobra.Command{
	Use:   "state <issue-id>",
	Short: "Show an issue's current state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var issue map[string]any
		if err := get("issues/"+args[0], &issue); err != nil {
			return err
		}
		fmt.Println(issue["state"])
		return nil
	},
}

var issuesStateUpdateCmd = &cobra.Command{
	Use:   "state-update <issue-id> <new-state>",
	Short: "Update an issue's state",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		newState := args[1]

		// Try as numeric ID first, otherwise treat as slug
		var issue map[string]any
		if _, err := strconv.Atoi(issueID); err == nil {
			if err := put("issues/"+issueID+"/state", map[string]string{"state": newState}, &issue); err != nil {
				return err
			}
		} else {
			// Resolve issue slug to ID via listing all projects' issues
			// Simpler approach: let the server resolve the slug
			if err := put("issues/"+issueID+"/state", map[string]string{"state": newState}, &issue); err != nil {
				return err
			}
		}

		b, _ := json.MarshalIndent(issue, "", "  ")
		fmt.Println(string(b))
		return nil
	},
}

// resolveProject tries to resolve a project slug to its numeric ID.
// Returns the ID as a string suitable for API paths.
func resolveProject(identifier string) string {
	// If it looks numeric, use it as-is
	if _, err := strconv.Atoi(identifier); err == nil {
		return identifier
	}
	// Otherwise, list all projects and match by slug
	var projects []struct {
		ID   int64  `json:"id"`
		Slug string `json:"slug"`
	}
	if err := get("projects", &projects); err != nil {
		return ""
	}
	for _, p := range projects {
		if p.Slug == identifier {
			return strconv.FormatInt(p.ID, 10)
		}
	}
	return ""
}

func init() {
	issuesCmd.AddCommand(issuesListCmd)
	issuesCmd.AddCommand(issuesShowCmd)
	issuesCmd.AddCommand(issuesCreateCmd)
	issuesCmd.AddCommand(issuesStateCmd)
	issuesCmd.AddCommand(issuesStateUpdateCmd)
	issuesCmd.AddCommand(issuesUpdateCmd)

	issuesListCmd.Flags().StringP("state", "s", "", "Filter by state")
	issuesListCmd.Flags().StringP("assignee", "a", "", "Filter by assignee ID")

	issuesCreateCmd.Flags().StringP("description", "d", "", "Issue description")
	issuesCreateCmd.Flags().StringP("type", "", "feature", "Issue type (epic, feature, bug, chore)")
	issuesCreateCmd.Flags().StringP("state", "", "todo", "Initial state")
	issuesCreateCmd.Flags().Int64P("assignee", "a", 0, "Assignee user ID")
	issuesCreateCmd.Flags().IntP("priority", "p", 3, "Priority (0=none, 1=urgent, 2=high, 3=medium, 4=low)")

	issuesUpdateCmd.Flags().StringP("title", "t", "", "New title")
	issuesUpdateCmd.Flags().StringP("description", "d", "", "New description")
	issuesUpdateCmd.Flags().StringP("type", "", "", "New type (epic, feature, bug, chore)")
	issuesUpdateCmd.Flags().StringP("state", "s", "", "New state")
	issuesUpdateCmd.Flags().Int64P("assignee", "a", 0, "New assignee user ID")
	issuesUpdateCmd.Flags().IntP("priority", "p", 0, "New priority")
}
