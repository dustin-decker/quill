package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/anchore/clio"
	"github.com/dustin-decker/quill/internal/bus"
	"github.com/dustin-decker/quill/quill/pki/apple"
)

func EmbeddedCerts(app clio.Application) *cobra.Command {
	return app.SetupCommand(&cobra.Command{
		Aliases: []string{
			"embedded-certs",
		},
		Use:   "embedded-certificates",
		Short: "show the certificates embedded into quill (typically the Apple root and intermediate certs)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer bus.Exit()

			var err error
			buf := &strings.Builder{}

			err = showAppleCerts(buf)

			if err != nil {
				return err
			}

			bus.Report(buf.String())

			return nil
		},
	})
}

func showAppleCerts(buf io.Writer) error {
	store := apple.GetEmbeddedCertStore()

	for _, cert := range store.RootPEMs() {
		if _, err := buf.Write([]byte(fmt.Sprintln(string(cert)))); err != nil {
			return fmt.Errorf("unable to write certificate: %w", err)
		}
	}

	for _, cert := range store.IntermediatePEMs() {
		if _, err := buf.Write([]byte(fmt.Sprintln(string(cert)))); err != nil {
			return fmt.Errorf("unable to write certificate: %w", err)
		}
	}

	return nil
}
