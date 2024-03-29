/*
Package cmd implements the subcommands that the executable understands.
*/
package cmd

import (
	"io"

	"github.com/Unquabain/ephemeral/data"
	"github.com/Unquabain/ephemeral/envelope"
	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var receiveData struct {
	privateRequestFile string
	responseFile       string
	secretFile         string
}

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive a secret response to a request for secret information.",
	Long: `Pairs the secret request (generated by the request subcommand)
with the encrypted response (generated by the respond subcommand) and
decrypts the secret.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			requestEnvelope, responseEnvelope envelope.Envelope
			requestFile, responseFile         io.ReadCloser
			secretFile                        io.WriteCloser
			request                           data.PrivateRequest
			response                          data.Response
			err                               error
		)
		if requestFile, err = openInputFile(receiveData.privateRequestFile); err != nil {
			log.WithError(err).Fatal(`Could not open private request file.`)
		}
		defer requestFile.Close()

		if _, err := requestEnvelope.ReadFrom(requestFile); err != nil {
			log.WithError(err).Fatal(`Could not read private request file.`)
		}
		if err := requestEnvelope.Open(&request); err != nil {
			log.WithError(err).Fatal(`Could not open private request envelope.`)
		}

		if responseFile, err = openInputFile(receiveData.responseFile); err != nil {
			log.WithError(err).Fatal(`Could not open response file.`)
		}
		defer responseFile.Close()

		secretFile, err = openOutputFile(receiveData.secretFile)
		if err != nil {
			log.WithError(err).Fatal(`Could not open secret file.`)
		}
		defer secretFile.Close()

		if _, err := responseEnvelope.ReadFrom(responseFile); err != nil {
			log.WithError(err).Fatal(`Could not read private response file.`)
		}
		if err := responseEnvelope.Open(&response); err != nil {
			log.WithError(err).Fatal(`Could not open response envelope.`)
		}
		if secret, err := request.Decode(response); err != nil {
			log.WithError(err).Fatal(`Could not decode secret.`)
		} else if _, err := secretFile.Write(secret); err != nil {
			log.WithError(err).Fatal(`Could not write secret file.`)
		}
	},
}

func init() {
	rootCmd.AddCommand(receiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// receiveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// receiveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	receiveCmd.Flags().StringVarP(&receiveData.privateRequestFile, `private`, `v`, `request_private.txt`, "The name of the private request file to be used to decode the response.")
	receiveCmd.Flags().StringVarP(&receiveData.responseFile, `response`, `r`, `-`, "The file the response was written to.")
	receiveCmd.Flags().StringVarP(&receiveData.secretFile, `secret`, `s`, `-`, "Where to write the decrypted, secret data.")
}
