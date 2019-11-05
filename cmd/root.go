package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// K8sEventListenerCommand main application
type K8sEventListenerCommand struct {
	rootCommand   *cobra.Command
	eventListener *EventListener
}

// NewK8sEventListenerCommand returns a pointer to K8sEventListenerCommand
func NewK8sEventListenerCommand() *K8sEventListenerCommand {
	return &K8sEventListenerCommand{
		rootCommand: getRootCommand(),
	}
}

// Run the main application
func (k *K8sEventListenerCommand) Run() int {
	/*k.rootCommand.Flags().StringSliceP("namespace", "n", nil, "K8s namespace")
	k.rootCommand.Flags().StringSliceP("label", "l", nil, "K8s endpoint matching label")
	k.rootCommand.Flags().StringSliceP("port-name", "p", nil, "K8s endpoint matching port name")
	k.rootCommand.Flags().Duration("timeout", 5*time.Second, "Proxy timeout")*/

	k.rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		k.rootCommand.Flags().VisitAll(bindFlags)

		k.eventListener = NewEventListener(
			viper.GetString("kube_config"),
			viper.GetString("kube_context"),
			k.handleError,
		)

		return k.eventListener.Init()
	}

	k.rootCommand.RunE = func(cmd *cobra.Command, args []string) (err error) {
		return k.eventListener.Listen()
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

	return
}
