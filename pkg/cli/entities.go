package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"webuntis-go-cli/pkg/api"
	"webuntis-go-cli/pkg/config"
)

// Helper to get authenticated client. Automatically re-authenticates
// using stored credentials when the session is missing or expired.
func getClient() (*api.Client, error) {
	cfg, accName := config.GetActiveAccount()
	if accName == "" {
		return nil, fmt.Errorf("no account found. Run 'webuntis login' first")
	}

	hasSession := cfg.SessionID != ""
	hasJWT := cfg.JWTToken != ""

	if hasSession || hasJWT {
		if hasJWT && isJWTExpired(cfg.JWTToken) && cfg.Username != "" && cfg.Password != "" {
			fmt.Fprintf(os.Stderr, "JWT expired, re-authenticating as '%s'...\n", cfg.Username)
			freshClient := api.NewClient(cfg.Server, cfg.School)
			if _, err := freshClient.Authenticate(cfg.Username, cfg.Password); err != nil {
				return nil, fmt.Errorf("auto re-authentication failed: %w", err)
			}
			cfg.SessionID = freshClient.SessionID
			cfg.JWTToken = freshClient.JWTToken
			cfg.TenantID = freshClient.TenantID
			cfg.SchoolYearID = freshClient.SchoolYearID
			cfg.CSRFToken = freshClient.CSRFToken
			if cfg.ActiveStudentID == 0 && len(freshClient.Students) > 0 {
				cfg.ActiveStudentID = freshClient.Students[0].ID
				cfg.ActiveStudentName = freshClient.Students[0].DisplayName
			}
			if err := config.AddOrUpdateAccount(accName, cfg, false); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save refreshed session: %v\n", err)
			}
			return freshClient, nil
		}

		client := api.NewClient(cfg.Server, cfg.School)
		client.SessionID = cfg.SessionID
		client.JWTToken = cfg.JWTToken
		client.TenantID = cfg.TenantID
		client.SchoolYearID = cfg.SchoolYearID
		client.CSRFToken = cfg.CSRFToken
		return client, nil
	}

	// No session — attempt auto-re-authentication with stored credentials
	if cfg.Username == "" || cfg.Password == "" {
		return nil, fmt.Errorf("session expired and no stored credentials. Run 'webuntis login' first")
	}

	fmt.Fprintf(os.Stderr, "Session expired, re-authenticating as '%s'...\n", cfg.Username)

	client := api.NewClient(cfg.Server, cfg.School)
	if _, err := client.Authenticate(cfg.Username, cfg.Password); err != nil {
		return nil, fmt.Errorf("auto re-authentication failed: %w", err)
	}

	// Persist the fresh tokens
	cfg.SessionID = client.SessionID
	cfg.JWTToken = client.JWTToken
	cfg.TenantID = client.TenantID
	cfg.SchoolYearID = client.SchoolYearID
	cfg.CSRFToken = client.CSRFToken

	if cfg.ActiveStudentID == 0 && len(client.Students) > 0 {
		cfg.ActiveStudentID = client.Students[0].ID
		cfg.ActiveStudentName = client.Students[0].DisplayName
		fmt.Fprintf(os.Stderr, "Active student set to '%s' (ID: %d)\n", cfg.ActiveStudentName, cfg.ActiveStudentID)
	}

	if err := config.AddOrUpdateAccount(accName, cfg, false); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save refreshed session: %v\n", err)
	}

	return client, nil
}

func isJWTExpired(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return true
	}
	return time.Now().Unix() >= claims.Exp
}

var classesCmd = &cobra.Command{
	Use:   "classes",
	Short: "List all classes",
	Long: `List all classes.

Available fields for --fields: id, name, longName, active, did`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		classes, err := client.GetKlassen()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		printJSON(classes)
	},
}

var teachersCmd = &cobra.Command{
	Use:   "teachers",
	Short: "List all teachers",
	Long: `List all teachers.

Available fields for --fields: id, name, longName, active, title, foreName, surName`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		teachers, err := client.GetTeachers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		printJSON(teachers)
	},
}

var subjectsCmd = &cobra.Command{
	Use:   "subjects",
	Short: "List all subjects",
	Long: `List all subjects.

Available fields for --fields: id, name, longName, active, alternateName`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		subjects, err := client.GetSubjects()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		printJSON(subjects)
	},
}

var roomsCmd = &cobra.Command{
	Use:   "rooms",
	Short: "List all rooms",
	Long: `List all rooms.

Available fields for --fields: id, name, longName, active, building`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		rooms, err := client.GetRooms()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		printJSON(rooms)
	},
}

func init() {
	rootCmd.AddCommand(classesCmd)
	rootCmd.AddCommand(teachersCmd)
	rootCmd.AddCommand(subjectsCmd)
	rootCmd.AddCommand(roomsCmd)
}
