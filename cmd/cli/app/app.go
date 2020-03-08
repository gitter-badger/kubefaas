package app

import (
	"github.com/spf13/cobra"

	"github.com/srcmesh/kubefaas/pkg/controller/client"
	"github.com/srcmesh/kubefaas/pkg/controller/client/rest"
	"github.com/srcmesh/kubefaas/pkg/cli/cliwrapper/cli"
	wrapper "github.com/srcmesh/kubefaas/pkg/cli/cliwrapper/driver/cobra"
	"github.com/srcmesh/kubefaas/pkg/cli/cliwrapper/driver/cobra/helptemplate"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/canaryconfig"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/environment"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/function"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/httptrigger"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/kubewatch"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/mqtrigger"
	_package "github.com/srcmesh/kubefaas/pkg/cli/cmd/package"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/spec"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/support"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/timetrigger"
	"github.com/srcmesh/kubefaas/pkg/cli/cmd/version"
	"github.com/srcmesh/kubefaas/pkg/cli/console"
	"github.com/srcmesh/kubefaas/pkg/cli/flag"
	flagkey "github.com/srcmesh/kubefaas/pkg/cli/flag/key"
	"github.com/srcmesh/kubefaas/pkg/cli/util"
	_ "github.com/srcmesh/kubefaas/pkg/mqtrigger/messageQueue/azurequeuestorage"
	_ "github.com/srcmesh/kubefaas/pkg/mqtrigger/messageQueue/kafka"
	_ "github.com/srcmesh/kubefaas/pkg/mqtrigger/messageQueue/nats"
)

const (
	usage = `Kubefaas: Fast and Simple Serverless Functions for Kubernetes

 * Github: https://github.com/srcmesh/kubefaas
`
)

func App() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:  "kubefaas",
		Long: usage,
		//SilenceUsage: true,
		PersistentPreRunE: wrapper.Wrapper(
			func(input cli.Input) error {
				console.Verbosity = input.Int(flagkey.Verbosity)

				if input.IsSet(flagkey.ClientOnly) {
					// TODO: use fake rest client for offline spec generation
					cmd.SetClientset(client.MakeFakeClientset(nil))
				} else {
					serverUrl, err := util.GetServerURL(input)
					if err != nil {
						return err
					}
					restClient := rest.NewRESTClient(serverUrl)
					cmd.SetClientset(client.MakeClientset(restClient))
				}

				return nil
			},
		),
	}

	// Workaround fix for not to show help command
	// https://github.com/spf13/cobra/issues/587
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	wrapper.SetFlags(rootCmd, flag.FlagSet{
		Global: []flag.Flag{flag.GlobalServer, flag.GlobalVerbosity},
	})

	groups := helptemplate.CommandGroups{}
	groups = append(groups, helptemplate.CreateCmdGroup("Basic Commands", environment.Commands(), _package.Commands(), function.Commands()))
	groups = append(groups, helptemplate.CreateCmdGroup("Trigger Commands", httptrigger.Commands(), mqtrigger.Commands(), timetrigger.Commands(), kubewatch.Commands()))
	groups = append(groups, helptemplate.CreateCmdGroup("Deploy Strategies Commands", canaryconfig.Commands()))
	groups = append(groups, helptemplate.CreateCmdGroup("Declarative Application Commands", spec.Commands()))
	groups = append(groups, helptemplate.CreateCmdGroup("Other Commands", support.Commands(), version.Commands()))
	groups.Add(rootCmd)

	flagExposer := helptemplate.ActsAsRootCommand(rootCmd, nil, groups...)
	// show global options in usage
	flagExposer.ExposeFlags(rootCmd, flagkey.Server, flagkey.Verbosity)

	return rootCmd
}
