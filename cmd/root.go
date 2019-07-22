package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"

	"github.com/hawari17/page-profiler/network"
)

var outputPath string
var configPath string

func init() {
	rootCmd.LocalNonPersistentFlags().StringVarP(&outputPath, "output-path", "o", "output.txt", "Specify directory Path path for output")
	rootCmd.LocalNonPersistentFlags().StringVarP(&configPath, "config-path", "c", "config.toml", "Path to configuration file")
}

var rootCmd = &cobra.Command{
	Use:   "page-profile",
	Short: "Page Profile is a tool to analyze web page using Chrome DevTools",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pageURL := args[0]
		if !govalidator.IsURL(pageURL) {
			log.Fatalf("%s is not a valid URL\n", pageURL)
			os.Exit(1)
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
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		network.MonitorPageNetwork(ctx, outFile, pageURL)
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
