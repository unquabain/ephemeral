package cmd

import (
	"io"

	"github.com/Unquabain/ephemeral/data"
	"github.com/Unquabain/ephemeral/envelope"
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var requestData struct {
	privateRequestFile string
	publicRequestFile  string
	description        string
}

// requestCmd represents the request command
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Create a request for secret data.",
	Long: `Creates two files: a public request that can be shared over public
channels, and a private request, which will be used to decode the response.
`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			privateEnvelope, publicEnvelope envelope.Envelope
			privateFile, publicFile         io.WriteCloser
		)
		request, err := data.NewRequest(requestData.description)
		if err != nil {
			log.WithError(err).Fatal(`Could not create a new request.`)
		}
		privateFile, err = openOutputFile(requestData.privateRequestFile)
		if err != nil {
			log.WithError(err).Fatal(`Could not open private request file.`)
		}
		defer privateFile.Close()
		publicFile, err = openOutputFile(requestData.publicRequestFile)
		if err != nil {
			log.WithError(err).Fatal(`Could not open public request file.`)
		}
		defer publicFile.Close()
		privateEnvelope.Name = `PRIVATE REQUEST`
		privateEnvelope.Prelude = request.Description
		if err := privateEnvelope.Stuff(request); err != nil {
			log.WithError(err).Fatal(`Could not encode private request.`)
		}

		if _, err := io.Copy(privateFile, privateEnvelope.Reader()); err != nil {
			log.WithError(err).Fatal(`Could not write request to private request file`)
		}

		publicEnvelope.Name = `PUBLIC REQUEST`
		publicEnvelope.Prelude = request.Description
		if err := publicEnvelope.Stuff(request.Public()); err != nil {
			log.WithError(err).Fatal(`Could not write encode public request.`)
		}
		if _, err := io.Copy(publicFile, publicEnvelope.Reader()); err != nil {
			log.WithError(err).Fatal(`Could not write request to public request file.`)
		}
	},
}

func init() {
	rootCmd.AddCommand(requestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// requestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	requestCmd.Flags().StringVarP(&requestData.privateRequestFile, `private`, `v`, `request_private.txt`, "The name of the secret request file to be used to decode the response.")
	requestCmd.Flags().StringVarP(&requestData.publicRequestFile, `public`, `b`, `-`, "The name of the public request file to be sent over public channels.")
	requestCmd.Flags().StringVarP(&requestData.description, `description`, `d`, `Secret Information`, "An optional description of the secret being requested.")
}
