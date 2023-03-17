package cmd

import (
	"fmt"

	"github.com/mudkipme/inazuma/config"
	"github.com/mudkipme/inazuma/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Inazuma proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		server.StartServer(conf)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
