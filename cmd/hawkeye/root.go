package hawkeye

import (
 "fmt"
 "os"

 "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:  "hawkeye",
    Short: "hawkeye - a tool to diagnose the Portworx cluster",
    Long: `hawkeye is a tool to diagnose the Portworx cluster

One can use hawkeye to debug the problems in the PX cluster`,
    Run: func(cmd *cobra.Command, args []string) {

    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
        os.Exit(1)
    }
}