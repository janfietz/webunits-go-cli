package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/janfietz/webunits-go-cli/pkg/config"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage multiple accounts",
	Long:  `List, switch, and manage WebUntis accounts.`,
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts",
	Run: func(cmd *cobra.Command, args []string) {
		accounts := config.ListAccounts()
		_, current := config.GetActiveAccount()
		
		if len(accounts) == 0 {
			fmt.Println("No accounts found.")
			return
		}

		for _, name := range accounts {
			prefix := "  "
			if name == current {
				prefix = "* "
			}
			fmt.Printf("%s%s\n", prefix, name)
		}
	},
}

var accountSwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch active account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		err := config.SwitchAccount(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Switched to account '%s'\n", name)
	},
}

var accountCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current account details",
	Run: func(cmd *cobra.Command, args []string) {
		acc, name := config.GetActiveAccount()
		if name == "" {
			fmt.Println("No active account.")
			return
		}
		
		fmt.Printf("Account: %s\n", name)
		fmt.Printf("Server:  %s\n", acc.Server)
		fmt.Printf("School:  %s\n", acc.School)
		fmt.Printf("User:    %s\n", acc.Username)
		if acc.SessionID != "" {
			fmt.Println("Status:  Logged In")
		} else {
			fmt.Println("Status:  Logged Out")
		}
	},
}

var accountDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		err := config.DeleteAccount(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted account '%s'\n", name)
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
	accountCmd.AddCommand(accountListCmd)
	accountCmd.AddCommand(accountSwitchCmd)
	accountCmd.AddCommand(accountCurrentCmd)
	accountCmd.AddCommand(accountDeleteCmd)
}
