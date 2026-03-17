package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"webuntis-go-cli/pkg/api"
	"webuntis-go-cli/pkg/config"
)

var (
	classID    int
	teacherID  int
	roomID     int
	subjectID  int
	studentID  int
	dateStr    string
	endDateStr string
)

var timetableCmd = &cobra.Command{
	Use:   "timetable",
	Short: "Get timetable entries",
	Long: `Get timetable entries for a specific class, teacher, room, subject, or student.
If no entity flag is given, the active student is used automatically.

Available fields for --fields: date, start, end, subject, subjectShort, teacher, teacherShort, room, status, substitutionText, lessonText, color`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		// Determine element type and ID
		var elemID int
		var resourceType string

		if classID != 0 {
			elemID = classID
			resourceType = "CLASS"
		} else if teacherID != 0 {
			elemID = teacherID
			resourceType = "TEACHER"
		} else if subjectID != 0 {
			elemID = subjectID
			resourceType = "SUBJECT"
		} else if roomID != 0 {
			elemID = roomID
			resourceType = "ROOM"
		} else if studentID != 0 {
			elemID = studentID
			resourceType = "STUDENT"
		} else {
			// Fall back to active student from config
			cfg := config.GetConfig()
			if cfg.ActiveStudentID != 0 {
				elemID = cfg.ActiveStudentID
				resourceType = "STUDENT"
			} else {
				fmt.Fprintf(os.Stderr, "Error: You must specify one of --class, --teacher, --subject, --room, --student, or set an active student with 'webuntis student set <id>'\n")
				os.Exit(1)
			}
		}

		// Parse dates as YYYY-MM-DD strings
		startDate := parseDateStr(dateStr)
		endDate := startDate
		if endDateStr != "" {
			endDate = parseDateStr(endDateStr)
		}

		resp, err := client.GetTimetableREST(resourceType, elemID, startDate, endDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		entries := flattenTimetable(resp)
		printJSON(entries)
	},
}

// parseDateStr returns a YYYY-MM-DD string. If input is empty, returns today.
func parseDateStr(d string) string {
	if d == "" {
		return time.Now().Format("2006-01-02")
	}
	// Validate the date parses correctly
	_, err := time.Parse("2006-01-02", d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing date '%s': %v\n", d, err)
		os.Exit(1)
	}
	return d
}

// extractTime extracts HH:MM from an ISO datetime string like "2026-03-09T08:30"
func extractTime(isoDateTime string) string {
	if idx := strings.Index(isoDateTime, "T"); idx >= 0 {
		return isoDateTime[idx+1:]
	}
	return isoDateTime
}

// flattenTimetable converts the nested REST response into flat entries
func flattenTimetable(resp *api.RestTimetableResponse) []api.FlatTimetableEntry {
	var entries []api.FlatTimetableEntry
	for _, day := range resp.Days {
		for _, entry := range day.GridEntries {
			flat := api.FlatTimetableEntry{
				Date:             day.Date,
				Start:            extractTime(entry.Duration.Start),
				End:              extractTime(entry.Duration.End),
				Status:           entry.Status,
				SubstitutionText: entry.SubstitutionText,
				LessonText:       entry.LessonText,
				Color:            entry.Color,
			}
			// Teacher from position1
			if len(entry.Position1) > 0 && entry.Position1[0].Current != nil {
				flat.Teacher = entry.Position1[0].Current.LongName
				flat.TeacherShort = entry.Position1[0].Current.ShortName
			} else if len(entry.Position1) > 0 && entry.Position1[0].Removed != nil {
				flat.Teacher = entry.Position1[0].Removed.LongName + " (cancelled)"
				flat.TeacherShort = entry.Position1[0].Removed.ShortName
			}
			// Subject from position2
			if len(entry.Position2) > 0 && entry.Position2[0].Current != nil {
				flat.Subject = entry.Position2[0].Current.LongName
				flat.SubjectShort = entry.Position2[0].Current.ShortName
			}
			// Room from position3
			if len(entry.Position3) > 0 && entry.Position3[0].Current != nil {
				flat.Room = entry.Position3[0].Current.ShortName
			}
			entries = append(entries, flat)
		}
	}
	return entries
}

func init() {
	rootCmd.AddCommand(timetableCmd)

	timetableCmd.Flags().IntVar(&classID, "class", 0, "Class ID")
	timetableCmd.Flags().IntVar(&teacherID, "teacher", 0, "Teacher ID")
	timetableCmd.Flags().IntVar(&roomID, "room", 0, "Room ID")
	timetableCmd.Flags().IntVar(&subjectID, "subject", 0, "Subject ID")
	timetableCmd.Flags().IntVar(&studentID, "student", 0, "Student ID")

	timetableCmd.Flags().StringVar(&dateStr, "date", "", "Date (YYYY-MM-DD), default today")
	timetableCmd.Flags().StringVar(&endDateStr, "end-date", "", "End Date (YYYY-MM-DD), default same as date")
}
