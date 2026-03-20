package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/janfietz/webunits-go-cli/pkg/config"
)

var studentsCmd = &cobra.Command{
	Use:   "students",
	Short: "List accessible students",
	Long: `List all students accessible to the logged-in account.
For LEGAL_GUARDIAN accounts, this returns the list of children.

Available fields for --fields: id, displayName`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		students, err := client.GetStudents()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		printJSON(students)
	},
}

var studentCmd = &cobra.Command{
	Use:   "student",
	Short: "Manage active student",
	Long:  `Manage the active student used for timetable and other student-specific commands.`,
}

var studentCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show active student",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()
		if cfg.ActiveStudentID == 0 {
			fmt.Println("No active student set. Run 'webuntis student set <id>' or re-login.")
			return
		}
		fmt.Printf("Student: %s\n", cfg.ActiveStudentName)
		fmt.Printf("ID:      %d\n", cfg.ActiveStudentID)
	},
}

var studentSetCmd = &cobra.Command{
	Use:   "set <student-id>",
	Short: "Set active student",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid student ID '%s'\n", args[0])
			os.Exit(1)
		}

		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		students, err := client.GetStudents()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		var found bool
		var name string
		for _, s := range students {
			if s.ID == id {
				found = true
				name = s.DisplayName
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Error: student ID %d not found. Available students:\n", id)
			for _, s := range students {
				fmt.Fprintf(os.Stderr, "  - %s (ID: %d)\n", s.DisplayName, s.ID)
			}
			os.Exit(1)
		}

		cfg, accName := config.GetActiveAccount()
		cfg.ActiveStudentID = id
		cfg.ActiveStudentName = name

		if err := config.AddOrUpdateAccount(accName, cfg, false); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Active student set to '%s' (ID: %d)\n", name, id)
	},
}

func init() {
	rootCmd.AddCommand(studentsCmd)
	rootCmd.AddCommand(studentCmd)
	studentCmd.AddCommand(studentCurrentCmd)
	studentCmd.AddCommand(studentSetCmd)
}
