package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"mini_frp2/modules/client"
	"mini_frp2/utils/log"
)

var frpcFileName string

var rootCmd = &cobra.Command{
	Use:   "frpc",
	Short: "The client of the frp",
	Run: func(cmd *cobra.Command, args []string) {
		err := client.LoadConf(frpcFileName)
		if err != nil {
			fmt.Printf("%s load failed: %v", frpcFileName, err)
			os.Exit(1)
		}

		log.InitLog(client.LogWay, client.LogFile, client.LogLevel)

		var wait sync.WaitGroup
		wait.Add(len(client.ClientProxys))
		for _, client := range client.ClientProxys {
			go ControlProcess(client, &wait)
		}
		log.Info("Start frpc success")

		wait.Wait()

		log.Warn("All proxy exit")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&frpcFileName, "config", "c", "", "Client Config file path")
	rootCmd.MarkFlagRequired("config")
}
