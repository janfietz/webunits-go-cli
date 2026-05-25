package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/janfietz/webunits-go-cli/pkg/api"
	"github.com/janfietz/webunits-go-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	creStudentID int
	creStartDate string
	creEndDate   string
)

var classRegEventsCmd = &cobra.Command{
	Use:   "classregevents",
	Short: "List class register entries (Klassenbucheinträge)",
	Long: `List class register entries (Klassenbucheinträge) for the active student or a specified student.

Each entry is either about an individual student (elemType=STUDENT) or the whole class (elemType=CLASS).

Available fields for --fields: date, time, elemType, elementName, subject, creator, reason, category, text`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		studentID := creStudentID
		if studentID == 0 {
			cfg := config.GetConfig()
			if cfg.ActiveStudentID != 0 {
				studentID = cfg.ActiveStudentID
			} else {
				fmt.Fprintf(os.Stderr, "Error: no student specified. Use --student <id> or set an active student with 'webuntis student set <id>'\n")
				os.Exit(1)
			}
		}

		start := creStartDate
		if start == "" {
			start = time.Now().AddDate(0, 0, -30).Format("20060102")
		} else {
			start = dateToYYYYMMDD(start)
		}
		end := creEndDate
		if end == "" {
			end = time.Now().Format("20060102")
		} else {
			end = dateToYYYYMMDD(end)
		}

		resp, err := client.GetClassRegEvents(studentID, start, end)
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		entries := flattenClassRegEvents(resp)
		printJSON(entries)
	},
}

func flattenClassRegEvents(resp *api.ClassRegEventsResponse) []api.FlatClassRegEvent {
	entries := make([]api.FlatClassRegEvent, 0, len(resp.Data.Rows))
	for _, e := range resp.Data.Rows {
		entries = append(entries, api.FlatClassRegEvent{
			Date:        formatIntDate(e.CreateDate),
			Time:        formatIntTime(e.CreateTime),
			ElemType:    e.ElemType,
			ElementName: e.ElementName,
			Subject:     e.SubjectName,
			Creator:     e.CreatorName,
			Reason:      e.EventReasonName,
			Category:    e.CategoryName,
			Text:        e.Text,
		})
	}
	return entries
}

func init() {
	rootCmd.AddCommand(classRegEventsCmd)
	classRegEventsCmd.Flags().IntVar(&creStudentID, "student", 0, "Student ID (overrides active student)")
	classRegEventsCmd.Flags().StringVar(&creStartDate, "start-date", "", "Start date (YYYY-MM-DD), default 30 days ago")
	classRegEventsCmd.Flags().StringVar(&creEndDate, "end-date", "", "End date (YYYY-MM-DD), default today")
}
