package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"

	"github.com/tiket-libre/ajax-detector/network"
)

var (
	outputPath string
	configPath string
	timeout    int
)

func init() {
	rootCmd.PersistentFlags().
		StringVarP(&outputPath, "output", "o", "output.csv", "Specify directory Path path for output")
	rootCmd.PersistentFlags().
		StringVarP(&configPath, "config", "c", "config.toml", "Path to configuration file")
	rootCmd.PersistentFlags().
		IntVarP(&timeout, "timeout", "t", 15, "Set timeout for the execution, in seconds")
}

var rootCmd = &cobra.Command{
	Use:   "page-profile",
	Short: "Page Profile is a tool to analyze web page using Chrome DevTools",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			if !cmd.Flags().Changed("config") {
				return errors.New("--config flag is required when providing no argument")
			}
			return nil
		}

		pageURLs := args[0:]
		for _, pageURL := range pageURLs {
			if !govalidator.IsURL(pageURL) {
				msg := fmt.Sprintf("%s is not a valid URL", pageURL)
				return errors.New(msg)
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var pages []string

		if cmd.Flags().Changed("config") {
			var err error
			if pages, err = readFromConfigFile(configPath); err != nil {
				log.Fatal(err)
			}
		} else {
			pages = args[0:]
		}

		outFile, err := createOutputFile(outputPath)
		if err != nil {
			log.Fatal(err)
		}

		// Create a timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		if err := network.LogAjaxRequest(ctx, outFile, pages); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute encapsulates the Execute method from rootCmd variable
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createOutputFile(filePath string) (io.Writer, error) {
	dirPath := path.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}

	return os.Create(filePath)
}

func readFromConfigFile(configPath string) ([]string, error) {
	pages := make([]string, 0)

	config, err := toml.LoadFile(configPath)
	if err != nil {
		return pages, err
	}

	pageConfigs := config.Get("pages").([]*toml.Tree)
	for _, pageConfig := range pageConfigs {
		pages = append(pages, pageConfig.Get("url").(string))
	}

	return pages, nil
}
