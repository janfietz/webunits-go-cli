package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/janfietz/webunits-go-cli/pkg/api"
	"github.com/janfietz/webunits-go-cli/pkg/config"
)

var (
	absStudentID int
	absStartDate string
	absEndDate   string
)

var absencesCmd = &cobra.Command{
	Use:   "absences",
	Short: "List student absences",
	Long: `List absences for the active student or a specified student.

Available fields for --fields: date, startTime, endTime, reason, text, studentName, excuseStatus, isExcused`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		// Determine student ID
		studentID := absStudentID
		if studentID == 0 {
			cfg := config.GetConfig()
			if cfg.ActiveStudentID != 0 {
				studentID = cfg.ActiveStudentID
			} else {
				fmt.Fprintf(os.Stderr, "Error: no student specified. Use --student <id> or set an active student with 'webuntis student set <id>'\n")
				os.Exit(1)
			}
		}

		// Default date range: last 30 days → today
		start := absStartDate
		if start == "" {
			start = time.Now().AddDate(0, 0, -30).Format("20060102")
		} else {
			start = dateToYYYYMMDD(start)
		}
		end := absEndDate
		if end == "" {
			end = time.Now().Format("20060102")
		} else {
			end = dateToYYYYMMDD(end)
		}

		resp, err := client.GetAbsences(studentID, start, end)
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		entries := flattenAbsences(resp)
		printJSON(entries)
	},
}

// dateToYYYYMMDD converts "YYYY-MM-DD" to "YYYYMMDD" for the API
func dateToYYYYMMDD(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing date '%s': %v\n", d, err)
		os.Exit(1)
	}
	return t.Format("20060102")
}

// formatIntDate converts YYYYMMDD int to "YYYY-MM-DD" string
func formatIntDate(d int) string {
	s := fmt.Sprintf("%d", d)
	if len(s) == 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:]
	}
	return s
}

// formatIntTime converts HHMM int to "HH:MM" string
func formatIntTime(t int) string {
	return fmt.Sprintf("%02d:%02d", t/100, t%100)
}

func flattenAbsences(resp *api.AbsencesResponse) []api.FlatAbsence {
	entries := make([]api.FlatAbsence, 0, len(resp.Data.Absences))
	for _, a := range resp.Data.Absences {
		entries = append(entries, api.FlatAbsence{
			Date:         formatIntDate(a.StartDate),
			StartTime:    formatIntTime(a.StartTime),
			EndTime:      formatIntTime(a.EndTime),
			Reason:       a.Reason,
			Text:         a.Text,
			StudentName:  a.StudentName,
			ExcuseStatus: a.ExcuseStatus,
			IsExcused:    a.IsExcused,
		})
	}
	return entries
}

func init() {
	rootCmd.AddCommand(absencesCmd)
	absencesCmd.Flags().IntVar(&absStudentID, "student", 0, "Student ID (overrides active student)")
	absencesCmd.Flags().StringVar(&absStartDate, "start-date", "", "Start date (YYYY-MM-DD), default 30 days ago")
	absencesCmd.Flags().StringVar(&absEndDate, "end-date", "", "End date (YYYY-MM-DD), default today")
}
