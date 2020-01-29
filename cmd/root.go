package cmd

import (
	"context"
	"fmt"
	"k8s-event-listener/pkg/eventlistener"
	"k8s-event-listener/pkg/resource"
	"log"
	"strings"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// K8sEventListenerCommand main application
type K8sEventListenerCommand struct {
	rootCommand   *cobra.Command
	eventListener *eventlistener.EventListener
	ctx           context.Context
	cErr          chan error
}

// NewK8sEventListenerCommand returns a pointer to K8sEventListenerCommand
func NewK8sEventListenerCommand(ctx context.Context) *K8sEventListenerCommand {
	return &K8sEventListenerCommand{
		rootCommand: getRootCommand(),
		ctx:         ctx,
		cErr:        make(chan error),
	}
}

// Run the main application
func (k *K8sEventListenerCommand) Run() int {
	k.rootCommand.Flags().StringP("resource", "r", "", "K8s resource to listen")
	k.rootCommand.Flags().StringP("callback", "c", "", "Callback to be executed")

	k.rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		k.rootCommand.Flags().VisitAll(bindFlags)

		k.eventListener = eventlistener.NewEventListener(
			k.ctx,
			viper.GetString("kube_config"),
			viper.GetString("kube_context"),
			func(err error) {
				k.cErr <- err
			},
			viper.GetString("verbose"),
		)

		return k.eventListener.Init()
	}

	k.rootCommand.RunE = func(cmd *cobra.Command, args []string) (err error) {
		r, err := resource.NewResource(viper.GetString("resource"), viper.GetString("callback"))
		if err != nil {
			return err
		}

		err = k.eventListener.Listen(r)
		if err != nil {
			return
		}

		select {
		case err = <-k.cErr:
			return
		}
	}

	if err := k.rootCommand.Execute(); err != nil {
		k.handleError(err)
		return 1
	}

	return 0
}

func (k *K8sEventListenerCommand) populateConfig() (err error) {
	viper.AddConfigPath(".")

	viper.SetConfigName(".config")
	viper.SetEnvPrefix("K8S_EVENT_LISTENER")
	viper.AutomaticEnv()

	return viper.ReadInConfig()
}

func (k *K8sEventListenerCommand) handleError(err error) {
	log.Println(fmt.Sprintf("[ERROR] %s",
		err.Error(),
	))
}

func bindFlags(flag *pflag.Flag) {
	if err := viper.BindPFlag(strings.ReplaceAll(flag.Name, "-", "_"), flag); err != nil {
		panic(err)
	}
}

func getRootCommand() (c *cobra.Command) {
	c = &cobra.Command{
		Use:           "k8s-event-listener",
		Short:         "Listen for specific kubernetes events",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	c.PersistentFlags().String("kube-config", "", "Path to kubeconfig file")
	c.PersistentFlags().String("kube-context", "", "Context to use")
	c.PersistentFlags().StringP("verbose", "v", "0", "Verbose level")

	return
}
