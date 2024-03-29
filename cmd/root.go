package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/multiconn"
	"github.com/overmindtech/overmind-agent/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "source-template",
	Short: "Remote primary source for kubernetes",
	Long: `A template for building sources.

Edit this once you have created your source
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get srcman supplied config
		natsServers := viper.GetStringSlice("nats-servers")
		natsNamePrefix := "agent"
		clientID := viper.GetString("client-id")
		clientSecret := viper.GetString("client-secret")
		overmindAuthURL := viper.GetString("overmind-auth-url")
		overmindTokenAPI := viper.GetString("overmind-token-api")
		maxParallel := viper.GetInt("max-parallel")
		startConnectRetries := viper.GetInt("start-connect-retries")
		hostname, err := os.Hostname()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not determine hostname for use in NATS connection name")

			os.Exit(1)
		}

		var clientSecretLog string

		if clientSecret != "" {
			clientSecretLog = "[REDACTED]"
		}

		log.WithFields(log.Fields{
			"nats-servers":          natsServers,
			"nats-name-prefix":      natsNamePrefix,
			"max-parallel":          maxParallel,
			"client-id":             clientID,
			"client-secret":         clientSecretLog,
			"overmind-auth-url":     overmindAuthURL,
			"start-connect-retries": startConnectRetries,
			"overmind-token-api":    overmindTokenAPI,
		}).Info("Got config")

		e := discovery.Engine{
			Name: "overmind-agent",
			NATSOptions: &multiconn.NATSConnectionOptions{
				CommonOptions: multiconn.CommonOptions{
					NumRetries: startConnectRetries,
					RetryDelay: 5 * time.Second,
				},
				Servers:           natsServers,
				ConnectionName:    fmt.Sprintf("%v.%v", natsNamePrefix, hostname),
				ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
				MaxReconnects:     30,                 // TODO: Make configurable
				ReconnectWait:     time.Second,        // TODO: Make configurable
				ReconnectJitter:   3 * time.Second,    // TODO: Make configurable
			},
			MaxParallelExecutions: maxParallel,
		}

		if clientID != "" && clientSecret != "" {
			log.WithFields(log.Fields{
				"client-id":          clientID,
				"client-secret":      clientSecretLog,
				"overmind-auth-url":  overmindAuthURL,
				"overmind-token-api": overmindTokenAPI,
			}).Info("Setting up authentication client")

			// Create a new token client to be used for NATS Auth. This with
			// authenticate to OAuth, then get a NATS-specific token from the
			// overmindTokenAPI
			oauthClient := multiconn.NewOAuthTokenClient(
				clientID,
				clientSecret,
				overmindAuthURL,
				overmindTokenAPI,
			)

			e.NATSOptions.TokenClient = oauthClient
		}

		// ⚠️ Here is where you add your sources
		e.AddSources(sources.Sources...)

		err = e.Start()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start engine")

			os.Exit(1)
		}

		sigs := make(chan os.Signal, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs

		log.Info("Stopping engine")

		err = e.Stop()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not stop engine")

			os.Exit(1)
		}

		log.Info("Stopped")

		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	var logLevel string

	home, _ := homedir.Dir()

	// General config options
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", fmt.Sprintf("%v/.overmind.yaml", home), "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")
	rootCmd.PersistentFlags().IntP("start-connect-retries", "r", 10, "How many time to try connecting on startup before giving up, set to -1 for infinite")

	// Config required by all sources in order to connect to NATS. You shouldn't
	// need to change these
	rootCmd.PersistentFlags().StringArray("nats-servers", []string{"nats://localhost:4222", "nats://nats:4222"}, "A list of NATS servers to connect to")
	rootCmd.PersistentFlags().String("nats-name-prefix", "", "A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client")
	rootCmd.PersistentFlags().Int("max-parallel", (runtime.NumCPU() * 2), "Max number of requests to run in parallel")

	// ⚠️ Add your own custom config options below, the example "your-custom-flag"
	// should be replaced with your own config or deleted
	rootCmd.PersistentFlags().String("client-id", "", "The client ID that will be used for authenticating with Overmind. Must be used in conjunction with --client-secret")
	rootCmd.PersistentFlags().String("client-secret", "", "Client secret associated with the supplied --client-id. Used to authenticate with Overmind")
	rootCmd.PersistentFlags().String("overmind-auth-url", "https://app.overmind.tech/todo/fix/this", "The URL to send Overmind authentication requests to")
	rootCmd.PersistentFlags().String("overmind-token-api", "https://app.overmind.tech/todo/v1", "The root URL of the overmind token API which is used to obtain NATS tokens")

	// Bind these to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				viper.BindPFlag(f.Name, f)
			}
		})
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	replacer := strings.NewReplacer("-", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}
}
