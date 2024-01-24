package cmd

import (
	"io"
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var privateRequestFile string

func openOutputFile(name string) (io.WriteCloser, error) {
	if name == `-` {
		return os.Stdout, nil
	}
	return os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
}
func openInputFile(name string) (io.ReadCloser, error) {
	if name == `-` {
		return os.Stdin, nil
	}
	return os.Open(name)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ephemeral",
	Short: "A way of requesting and sending secret info on public channels",
	Long: `Uses Elliptic Curve Diffie-Hellman key exchange to make a request
for secret information, and AES256 for responding to that request with
the secret.

First, a request is made. For example:
    ephemeral request -v secret.txt -b request.txt -d "The database password"
	
The file "request.txt" can be shared in Slack or another public channel. The
recipient saves it to a file, and can provide the information with:
    ephemeral respond -b request.txt -d "Swordfish" -r response.txt

Likewise, "response.txt" can be sent over public channels. Finally, the
original requester, who has saved "secret.txt" for this occasion, can
decrypt the message with:
    ephemeral receive -v secret.txt -r response.txt -s dbpassword.txt

STDIN and STDOUT can be used in conjunction with other programs, e.g.
"pbcopy" and "pbpaste" on Macs, to streamline the operation.

    ephemeral request -v secret.txt -d "Cluster Certificate" | pbcopy
		pbpaste | ephemeral respond -d cert.pem | pbcopy
		pbpaste | ephemeral -v secret.txt | pbcopy


`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().BoolP("debug", "g", false, "Turn on verbose logging.")
	cobra.OnInitialize(initLogging)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

func initLogging() {
	if debug := rootCmd.PersistentFlags().Lookup(`debug`); debug == nil {
		log.SetLevel(log.WarnLevel)
	} else if debug.Changed {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}
