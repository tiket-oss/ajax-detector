package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/chromedp/chromedp"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"

	"github.com/tiket-libre/ajax-detector/network"
)

type pageInfo struct {
	name string
	url  string
}

var outputPath string
var configPath string
var timeout int

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputPath, "output-path", "o", "output.txt", "Specify directory Path path for output")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", "config.toml", "Path to configuration file")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 15, "Set timeout for the execution, in seconds")
}

var rootCmd = &cobra.Command{
	Use:   "page-profile",
	Short: "Page Profile is a tool to analyze web page using Chrome DevTools",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pages := make([]pageInfo, 0)

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

			pages = append(pages, pageInfo{name: fmt.Sprintf("%s - Network", pageURL), url: pageURL})
		}

		outFile, err := createOutputFile(outputPath)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		ctx, cancel := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
		defer cancel()

		// Create a timeout
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		for _, page := range pages {
			network.MonitorPageNetwork(ctx, outFile, page.url)
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

func readFromConfigFile(configPath string) ([]pageInfo, error) {
	pages := make([]pageInfo, 0)

	config, err := toml.LoadFile(configPath)
	if err != nil {
		return pages, err
	}

	pageConfigs := config.Get("pages").([]*toml.Tree)
	for _, pageConfig := range pageConfigs {
		pages = append(pages, pageInfo{
			name: pageConfig.Get("name").(string),
			url:  pageConfig.Get("url").(string),
		})
	}

	return pages, nil
}
