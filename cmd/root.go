package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"

	ajaxdetector "github.com/tiket-libre/ajax-detector"
	"github.com/tiket-libre/ajax-detector/network"
)

var outputPath string
var configPath string
var timeout int

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output-path", "o", "output.csv", "Specify directory Path path for output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", "config.toml", "Path to configuration file")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 15, "Set timeout for the execution, in seconds")
}

var rootCmd = &cobra.Command{
	Use:   "page-profile",
	Short: "Page Profile is a tool to analyze web page using Chrome DevTools",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pages := make([]ajaxdetector.PageInfo, 0)

		if cmd.Flags().Changed("config-path") {
			var err error
			pages, err = readFromConfigFile(configPath)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		} else {
			pageURL := args[0]
			if pageURL != "" && !govalidator.IsURL(pageURL) {
				log.Fatalf("%s is not a valid URL\n", pageURL)
				os.Exit(1)
			}

			pages = append(pages, ajaxdetector.PageInfo{URL: pageURL})
		}

		outFile, err := createOutputFile(outputPath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
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
		log.Println(err)
		os.Exit(1)
	}
}

func createOutputFile(filePath string) (io.Writer, error) {
	dirPath := path.Dir(filePath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, os.ModePerm)
	}

	return os.Create(filePath)
}

func readFromConfigFile(configPath string) ([]ajaxdetector.PageInfo, error) {
	pages := make([]ajaxdetector.PageInfo, 0)

	config, err := toml.LoadFile(configPath)
	if err != nil {
		return pages, err
	}

	pageConfigs := config.Get("pages").([]*toml.Tree)
	for _, pageConfig := range pageConfigs {
		pages = append(pages, ajaxdetector.PageInfo{
			Name: pageConfig.Get("name").(string),
			URL:  pageConfig.Get("url").(string),
		})
	}

	return pages, nil
}
