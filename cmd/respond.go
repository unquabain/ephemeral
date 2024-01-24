package cmd

import (
	"bytes"
	"io"

	"github.com/Unquabain/ephemeral/data"
	"github.com/Unquabain/ephemeral/envelope"
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var respondData struct {
	publicRequestFile string
	dataFile          string
	responseFile      string
}

// respondCmd represents the respond command
var respondCmd = &cobra.Command{
	Use:   "respond",
	Short: "Reply to a request for secret information",
	Long: `If given a public request (generated with the request subcommand),
formulate a reply.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			requestEnvelope, responseEnvelope envelope.Envelope
			requestFile, dataFile             io.ReadCloser
			responseFile                      io.WriteCloser
			request                           data.PublicRequest
			err                               error
		)
		requestFile, err = openInputFile(respondData.publicRequestFile)
		if err != nil {
			log.WithError(err).Fatal(`Could not open request file.`)
		}
		defer requestFile.Close()

		if _, err := requestEnvelope.ReadFrom(requestFile); err != nil {
			log.WithError(err).Fatal(`Could not read request file: %s`)
		}
		if err := requestEnvelope.Open(&request); err != nil {
			log.WithError(err).Fatal(`Could not open request envelope: %s`)
		}

		dataFile, err = openInputFile(respondData.dataFile)
		if err != nil {
			log.WithError(err).Fatal(`Could not open data file: %s`)
		}
		defer dataFile.Close()

		responseFile, err = openOutputFile(respondData.responseFile)
		if err != nil {
			log.WithError(err).Fatal(`Could not open output file: %s`)
		}
		defer responseFile.Close()

		buff := new(bytes.Buffer)
		if _, err := io.Copy(buff, dataFile); err != nil {
			log.WithError(err).Fatal(`Could not read data file: %s`)
		}

		responseEnvelope.Name = `RESPONSE`
		responseEnvelope.Prelude = request.Description
		if response, err := request.Encode(buff.Bytes()); err != nil {
			log.WithError(err).Fatal(`Could not encode response: %s`)
		} else if err := responseEnvelope.Stuff(response); err != nil {
			log.WithError(err).Fatal(`Could not stuff response envelope: %s`)
		}
		if _, err := io.Copy(responseFile, responseEnvelope.Reader()); err != nil {
			log.WithError(err).Fatal(`Could not write response file: %s`)
		}
	},
}

func init() {
	rootCmd.AddCommand(respondCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// respondCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// respondCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	respondCmd.Flags().StringVarP(&respondData.publicRequestFile, `public`, `b`, `-`, "The name of the public request file sent over public channels.")
	respondCmd.Flags().StringVarP(&respondData.dataFile, `data`, `d`, `-`, "A data file to encrypt in the response.")
	respondCmd.Flags().StringVarP(&respondData.responseFile, `response`, `r`, `-`, "The file to write the response to.")
}
