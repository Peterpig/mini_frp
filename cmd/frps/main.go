package main

import (
	"fmt"
	"mini_frp2/modules/conn"
	"mini_frp2/modules/server"
	"mini_frp2/utils/log"
	"os"

	"github.com/spf13/cobra"
)

var frpsFileName string

var rootCmd = &cobra.Command{
	Use:   "frps",
	Short: "The server of the frp",
	Run: func(cmd *cobra.Command, args []string) {
		err := server.LoadConf(frpsFileName)
		if err != nil {
			fmt.Printf("%s load failed: %v", frpsFileName, err)
			os.Exit(1)
		}

		log.InitLog(server.LogWay, server.LogFile, server.LogLevel)

		l, err := conn.Linsten(server.BindAddr, server.BindPort)
		if err != nil {
			log.Error("Frp server start field : %v", err)
		}
		log.Info("Frp server start success: %s", l.Addr())
		ProcessControl(l)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&frpsFileName, "config", "c", "", "Server Config file path")
	rootCmd.MarkFlagRequired("config")
}
