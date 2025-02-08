package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wwwlde/quakejs-proxy/proxy"
)

var (
	wsServer   string
	listenAddr string
	hexdump    bool
	logNewConn bool
	logExch    bool
)

func main() {
	// Define the root command for the CLI
	var rootCmd = &cobra.Command{
		Use:   "quakejs-proxy",
		Short: "A proxy server for QuakeJS",
		Long:  `A UDP to WebSocket proxy server for QuakeJS.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Configure proxy settings based on flags
			proxy.SetHexdumpPackets(hexdump)
			proxy.SetLogExchanges(logExch)
			proxy.SetLogNewConnections(logNewConn)

			// Create and start the proxy server
			server := proxy.New(listenAddr, wsServer)
			defer server.Close()

			// Log server startup information
			logrus.WithFields(logrus.Fields{
				"listen": listenAddr,
				"dest":   wsServer,
			}).Info("Proxy server listening")

			// Start the server and handle errors
			if err := server.Start(); err != nil {
				logrus.WithField("err", err).Fatal("Failed to start proxy server")
			}
		},
	}

	// Define command-line flags
	rootCmd.Flags().StringVarP(&wsServer, "ws", "w", "127.0.0.1:27961", "Hostname of the WebSocket server")
	rootCmd.Flags().StringVarP(&listenAddr, "listen", "l", "", "Host to listen on")
	rootCmd.Flags().BoolVar(&hexdump, "hexdump", false, "Print a hex dump of every packet")
	rootCmd.Flags().BoolVar(&logNewConn, "log-new-conn", true, "Logs every new connection")
	rootCmd.Flags().BoolVar(&logExch, "log-exchanges", false, "Logs all exchanges going through the proxy")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		logrus.WithField("err", err).Fatal("Failed to execute command")
	}
}
