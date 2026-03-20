package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/janfietz/webunits-go-cli/pkg/api"
)

var (
	hwStartDate string
	hwEndDate   string
)

var homeworkCmd = &cobra.Command{
	Use:   "homework",
	Short: "List homework assignments",
	Long: `List homework assignments with due dates and completion status.

Available fields for --fields: date, dueDate, text, teacher, subject, remark, completed`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		start := hwStartDate
		if start == "" {
			start = time.Now().AddDate(0, 0, -30).Format("20060102")
		} else {
			start = dateToYYYYMMDD(start)
		}
		end := hwEndDate
		if end == "" {
			end = time.Now().Format("20060102")
		} else {
			end = dateToYYYYMMDD(end)
		}

		resp, err := client.GetHomework(start, end)
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		entries := flattenHomework(resp)
		printJSON(entries)
	},
}

func flattenHomework(resp *api.HomeworkResponse) []api.FlatHomework {
	teacherMap := make(map[int]string)
	for _, t := range resp.Data.Teachers {
		teacherMap[t.ID] = t.Name
	}
	lessonMap := make(map[int]string)
	for _, l := range resp.Data.Lessons {
		lessonMap[l.ID] = l.Subject
	}
	recordMap := make(map[int]int) // homeworkId → teacherId
	for _, r := range resp.Data.Records {
		recordMap[r.HomeworkID] = r.TeacherID
	}

	entries := make([]api.FlatHomework, 0, len(resp.Data.Homeworks))
	for _, hw := range resp.Data.Homeworks {
		flat := api.FlatHomework{
			Date:      formatIntDate(hw.Date),
			DueDate:   formatIntDate(hw.DueDate),
			Text:      hw.Text,
			Remark:    hw.Remark,
			Completed: hw.Completed,
		}
		if tid, ok := recordMap[hw.ID]; ok {
			flat.Teacher = teacherMap[tid]
		}
		flat.Subject = lessonMap[hw.LessonID]
		entries = append(entries, flat)
	}
	return entries
}

func init() {
	rootCmd.AddCommand(homeworkCmd)
	homeworkCmd.Flags().StringVar(&hwStartDate, "start-date", "", "Start date (YYYY-MM-DD), default 30 days ago")
	homeworkCmd.Flags().StringVar(&hwEndDate, "end-date", "", "End date (YYYY-MM-DD), default today")
}
