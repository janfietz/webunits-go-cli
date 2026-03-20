package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/janfietz/webunits-go-cli/pkg/api"
)

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "List inbox messages",
	Long: `List inbox messages from school.

Available fields for --fields: subject, sender, sentAt, contentPreview, isRead, hasAttachments`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := getClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
			os.Exit(ExitAuth)
		}

		resp, err := client.GetMessages()
		if err != nil {
			fmt.Fprintf(os.Stderr, "API Error: %v\n", err)
			os.Exit(ExitAPI)
		}

		entries := flattenMessages(resp)
		printJSON(entries)
	},
}

func flattenMessages(resp *api.MessagesResponse) []api.FlatMessage {
	entries := make([]api.FlatMessage, 0, len(resp.IncomingMessages))
	for _, m := range resp.IncomingMessages {
		entries = append(entries, api.FlatMessage{
			Subject:        m.Subject,
			Sender:         m.Sender.DisplayName,
			SentAt:         m.SentDateTime,
			ContentPreview: m.ContentPreview,
			IsRead:         m.IsMessageRead,
			HasAttachments: m.HasAttachments,
		})
	}
	return entries
}

func init() {
	rootCmd.AddCommand(messagesCmd)
}
