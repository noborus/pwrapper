package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/noborus/pwrapper/wrap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

var pwrapper wrap.PWrapper

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pwrapper",
	Short: "Wrap and execute command with pipe connected",
	Long:  `Wrap and execute command with pipe connected`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return wrap.Command(pwrapper)
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pwrapper.yaml)")
	rootCmd.PersistentFlags().StringVar(&pwrapper.WrapCommand, "wrap-command", "", "Command that receives standard input")
	rootCmd.PersistentFlags().StringVar(&pwrapper.Start, "start", "", "The first string to pipe to the command")
	rootCmd.PersistentFlags().StringVar(&pwrapper.End, "end", "", "The last string to pipe to the command")
	rootCmd.PersistentFlags().StringSliceVar(&pwrapper.ExecCommand, "exec", []string{}, "Command that receives standard input")
	rootCmd.PersistentFlags().BoolVar(&pwrapper.Debug, "debug", false, "enable debug print")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pwrapper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pwrapper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
