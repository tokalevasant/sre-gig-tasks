/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"strconv"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var Number, NumberNonblank bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gocat",
	Short: "A go equivalent of linux cat command",
	Long: `A go equivalent of linux cat command. 
Accepts following flags:
	-b, --number-nonblank
		number nonempty output lines, overrides -n
	
	-n, --number
		number all output lines
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		processArguments(args)
	},
	Args: cobra.MinimumNArgs(1),
}

func processArguments(args []string) {
	var lineNum uint = 0
	for _, arg := range args {
		fileInfo, err := os.Stat(arg)
		if err != nil {
			fmt.Printf("gocat: %s: No such file or directory\n", arg)
			return
		}

		if fileInfo.IsDir() { //print message below and continue - just like cat does
			fmt.Printf("cat: %s: Is a directory\n", arg)
			continue
		}
		lineNum = printFile(arg, lineNum)
	}
}

func printFile(arg string, lineNum uint) uint {
	file, err := os.Open(arg)
	if err != nil {
		fmt.Printf("gocat: %s: Permission denied\n", arg)
		return lineNum
	}
	defer file.Close()

	scanner := bufio.NewScanner(file) //scan the contents of a file and print line by line
	for scanner.Scan() {
		line := scanner.Text()

		onlySpaces := len(strings.TrimSpace(line)) == 0

		prefix := ""

		if (NumberNonblank && !onlySpaces) || (!NumberNonblank && Number) {
			lineNum++
			prefix = fmt.Sprintf("\t%5s\t", strconv.FormatUint(uint64(lineNum), 10))
		}

		fmt.Println(prefix, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from file:", err) //print error if scanning is not done properly
	}
	return lineNum
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gocat.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVarP(&Number, "number", "n", false, "number all output lines")
	rootCmd.Flags().BoolVarP(&NumberNonblank, "number-nonblank", "b", false, "number nonempty output lines, overrides -n")
}
