package hawkeye

import (
	"fmt"
	"log"

	"encoding/json"

	"github.com/nrevanna/hawk-eye/pkg/hawkeye"
	"github.com/spf13/cobra"
)

var argFocusObj string

var eventsCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"evt", "event"},
	Short:   "Outputs timeline of events in the PX cluster",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		events, err := hawkeye.GetEvents(args[0], argFocusObj)
		if err != nil {
			log.Fatalf("failed to get events: %v", err)
			return
		}
		b, err := json.Marshal(events)
		if err != nil {
			log.Fatalf("failed to marshal events: %v", err)
		}
		fmt.Println(string(b))
	},
}

func init() {
	eventsCmd.Flags().StringVarP(&argFocusObj, "focus", "f", "", "Focus on the specified object")
	rootCmd.AddCommand(eventsCmd)
}
