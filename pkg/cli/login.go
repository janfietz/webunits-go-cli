package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"webuntis-go-cli/pkg/api"
	"webuntis-go-cli/pkg/config"
)

var (
	server      string
	school      string
	username    string
	password    string
	accountName string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with WebUntis",
	Long:  `Authenticate with WebUntis API and store session token locally. Supports multiple accounts via --name.`,
	Run: func(cmd *cobra.Command, args []string) {
		promptMissing()

		client := api.NewClient(server, school)
		_, err := client.Authenticate(username, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
			os.Exit(ExitAuth)
		}

		// The client is now populated with all necessary tokens and IDs
		acc := config.Account{
			Server:       server,
			School:       school,
			Username:     username,
			Password:     password,
			SessionID:    client.SessionID,
			JWTToken:     client.JWTToken,
			TenantID:     client.TenantID,
			SchoolYearID: client.SchoolYearID,
			CSRFToken:    client.CSRFToken,
		}

		// Handle student selection for LEGAL_GUARDIAN accounts
		if len(client.Students) == 1 {
			acc.ActiveStudentID = client.Students[0].ID
			acc.ActiveStudentName = client.Students[0].DisplayName
			fmt.Fprintf(os.Stderr, "Active student set to '%s' (ID: %d)\n", acc.ActiveStudentName, acc.ActiveStudentID)
		} else if len(client.Students) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple students found. Please select:\n")
			for i, s := range client.Students {
				fmt.Fprintf(os.Stderr, "  %d) %s (ID: %d)\n", i+1, s.DisplayName, s.ID)
			}
			reader := bufio.NewReader(os.Stdin)
			fmt.Fprint(os.Stderr, "Select student [1]: ")
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			choice := 1
			if line != "" {
				fmt.Sscanf(line, "%d", &choice)
			}
			if choice < 1 || choice > len(client.Students) {
				choice = 1
			}
			acc.ActiveStudentID = client.Students[choice-1].ID
			acc.ActiveStudentName = client.Students[choice-1].DisplayName
			fmt.Fprintf(os.Stderr, "Active student set to '%s' (ID: %d)\n", acc.ActiveStudentName, acc.ActiveStudentID)
		}

		// Use provided name or default
		if accountName == "" {
			accountName = "default"
		}

		err = config.AddOrUpdateAccount(accountName, acc, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully logged in as '%s' (account: %s).\n", username, accountName)
		fmt.Printf("Tokens: SessionID=%v, JWT=%v, TenantID=%v, SchoolYearID=%v, CSRFToken=%v\n",
			acc.SessionID != "", acc.JWTToken != "", acc.TenantID != "", acc.SchoolYearID != "", acc.CSRFToken != "")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear session",
	Run: func(cmd *cobra.Command, args []string) {
		acc, name := config.GetActiveAccount()
		if acc.SessionID == "" {
			fmt.Println("Not logged in.")
			return
		}

		client := api.NewClient(acc.Server, acc.School)
		client.SessionID = acc.SessionID // Use the classic session for logout

		// Attempt logout
		if err := client.Logout(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: server logout failed: %v\n", err)
		}

		// Clear session locally but keep other details
		acc.SessionID = ""
		acc.JWTToken = ""
		acc.CSRFToken = ""

		err := config.AddOrUpdateAccount(name, acc, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Logged out successfully from account '%s'.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)

	loginCmd.Flags().StringVar(&server, "server", "", "WebUntis server domain (e.g. demo.webuntis.com)")
	loginCmd.Flags().StringVar(&school, "school", "", "School name")
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	loginCmd.Flags().StringVarP(&accountName, "name", "n", "default", "Account alias/name")

}

func promptMissing() {
	reader := bufio.NewReader(os.Stdin)

	if server == "" {
		server = promptLine(reader, "Server (e.g. demo.webuntis.com): ")
	}
	if school == "" {
		school = promptLine(reader, "School: ")
	}
	if username == "" {
		username = promptLine(reader, "Username: ")
	}
	if password == "" {
		fmt.Fprint(os.Stderr, "Password: ")
		raw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
			os.Exit(1)
		}
		password = string(raw)
	}
}

func promptLine(reader *bufio.Reader, prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
