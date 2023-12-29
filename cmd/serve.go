/*
Copyright © 2023 Mobile Technologies Inc. <connect-support@mtigs.com>
All Rights Reserved
*/
package cmd

import (
	"fmt"

	"github.com/Unquabain/ephemeral/server"
	"github.com/spf13/cobra"
)

var serveData struct {
	Addr string
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves a web server with a JSON RESTful API.",
	Long: `Listens on the address you specify, and offers three endpoints:
/request, /respond, and /receive, which correspond to the three subdommands.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Listening on %s. CTRL+C to stop\n", serveData.Addr)
		server.ListenAndServe(serveData.Addr)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().StringVarP(&serveData.Addr, "address", "a", ":8989", "Listen address.")
}
