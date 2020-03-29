/*
   Crossbar is a websocket relay
   Copyright (C) 2019 Timothy Drysdale <timothy.d.drysdale@gmail.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as
   published by the Free Software Foundation, either version 3 of the
   License, or (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var bufferSize int64
var cfgFile string
var cpuprofile string
var host *url.URL
var listen string
var logFile string
var development bool

/* configuration

bufferSize
muxBufferLength (for main message queue into the mux)
clientBufferLength (for each client's outgoing channel)

*/

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "shellbar",
	Short: "websocket relay with topics and exclusive connections",
	Long: `Shellbar is a websocket relay with topics set by the URL path, 
can handle binary and text messages, with one-to-one connections between 
servers (connecting to /server/<route>, speaking gob with per-connection ID) 
and clients (connecting to /<route>) speaking []byte.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		addr := viper.GetString("listen")
		development := viper.GetBool("development")

		if development {
			// development environment
			fmt.Printf("Devmode? %v\n", development)
			fmt.Printf("listening on %v\n", addr)
			log.SetFormatter(&log.TextFormatter{})
			log.SetLevel(log.TraceLevel)
			log.SetOutput(os.Stdout)
			fmt.Println("Setting log output to stdout")

		} else {

			//production environment
			log.SetFormatter(&log.JSONFormatter{})
			log.SetLevel(log.WarnLevel)

			file, err := os.OpenFile("shellbar.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				log.SetOutput(file)
				//fmt.Println("Setting log output to file")
			} else {
				log.Info("Failed to log to file, using default stderr")
			}

		}

		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		var wg sync.WaitGroup

		closed := make(chan struct{})

		c := make(chan os.Signal, 1)

		signal.Notify(c, os.Interrupt)

		go func() {
			for _ = range c {
				close(closed)
				wg.Wait()
				os.Exit(0)
			}
		}()

		wg.Add(1)

		go shellbar(addr, closed, &wg)

		wg.Wait()
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

	rootCmd.PersistentFlags().StringVarP(&listen, "listen", "l", "127.0.0.1:8080", "http://<ip>:<port> to listen on (default is 127.0.0.1:8080)")
	rootCmd.PersistentFlags().Int64VarP(&bufferSize, "buffer", "b", 32768, "bufferSize in bytes (default is 32,768)")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log", "f", "", "log file (default is STDOUT)")
	rootCmd.PersistentFlags().StringVarP(&cpuprofile, "cpuprofile", "c", "", "write cpu profile to file")
	rootCmd.PersistentFlags().BoolVarP(&development, "dev", "d", false, "development environment")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("SHELLBAR")
	viper.AutomaticEnv() // read in environment variables that match

}
