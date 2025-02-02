package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/anchore/clio"
	"github.com/dustin-decker/quill/cmd/quill/cli/options"
	"github.com/dustin-decker/quill/internal/bus"
	"github.com/dustin-decker/quill/internal/log"
	"github.com/dustin-decker/quill/quill"
	"github.com/dustin-decker/quill/quill/notary"
)

type submissionStatusConfig struct {
	ID             string `yaml:"id" json:"id" mapstructure:"-"`
	options.Notary `yaml:"notary" json:"notary" mapstructure:"notary"`
	options.Status `yaml:"status" json:"status" mapstructure:"status"`
}

func SubmissionStatus(app clio.Application) *cobra.Command {
	opts := &submissionStatusConfig{
		Status: options.Status{
			Wait: false,
		},
	}

	return app.SetupCommand(&cobra.Command{
		Use:   "status SUBMISSION_ID",
		Short: "check against Apple's Notary service to see the status of a notarization submission request",
		Example: options.FormatPositionalArgsHelp(
			map[string]string{
				"SUBMISSION_ID": "the submission ID to check the status of",
			},
		),
		Args: chainArgs(
			cobra.ExactArgs(1),
			func(_ *cobra.Command, args []string) error {
				opts.ID = args[0]
				return nil
			},
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer bus.Exit()

			log.Infof("checking submission status for %q", opts.ID)

			cfg := quill.NewNotarizeConfig(
				opts.Notary.Issuer,
				opts.Notary.PrivateKeyID,
				opts.Notary.PrivateKey,
			).WithStatusConfig(
				notary.StatusConfig{
					Timeout: time.Duration(int64(opts.TimeoutSeconds) * int64(time.Second)),
					Poll:    time.Duration(int64(opts.PollSeconds) * int64(time.Second)),
					Wait:    opts.Wait,
				},
			)

			token, err := notary.NewSignedToken(cfg.TokenConfig)
			if err != nil {
				return err
			}

			a := notary.NewAPIClient(token, cfg.HTTPTimeout)

			sub := notary.ExistingSubmission(a, opts.ID)

			var status notary.SubmissionStatus
			if opts.Wait {
				status, err = notary.PollStatus(cmd.Context(), sub, cfg.StatusConfig)
			} else {
				status, err = sub.Status(cmd.Context())
			}
			if err != nil {
				return err
			}

			bus.Report(string(status))

			return nil
		},
	}, opts)
}
