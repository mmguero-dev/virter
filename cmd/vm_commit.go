package cmd

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/LINBIT/virter/pkg/actualtime"
)

func vmCommitCommand() *cobra.Command {
	var shutdown bool

	var commitCmd = &cobra.Command{
		Use:   "commit name",
		Short: "Commit a virtual machine",
		Long: `Commit a virtual machine to an image. The image name will be
the virtual machine name.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			v, err := InitVirter()
			if err != nil {
				log.Fatal(err)
			}
			defer v.ForceDisconnect()

			shutdownTimeout := viper.GetDuration("time.shutdown_timeout")

			ctx, cancel := onInterruptWrap(context.Background())
			defer cancel()
			err = v.VMCommit(ctx, actualtime.ActualTime{}, args[0], shutdown, shutdownTimeout, viper.GetBool("libvirt.static_dhcp"))
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	commitCmd.Flags().BoolVarP(&shutdown, "shutdown", "s", false, "whether to shut the VM down and wait until it stops (default false)")

	return commitCmd
}
