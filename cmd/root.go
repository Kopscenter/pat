package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"path/filepath"
	"github.com/hashicorp/go-multierror"
	"log"
	pat "github.com/kevinjqiu/pat/pkg"
	"errors"
	"flag"
)

func collectTestFiles(globPatterns []string) ([]string, error) {
	if len(globPatterns) == 0 {
		return filepath.Glob("test_*.yaml")
	}

	var (
		err       error
		filePaths []string
	)

	for _, pattern := range globPatterns {
		files, newErr := filepath.Glob(pattern)
		if err != nil {
			err = multierror.Append(err, newErr)
		}

		for _, file := range files {
			if _, err := os.Stat(file); err == nil {
				filePaths = append(filePaths, file)
			}
		}
	}

	return filePaths, err
}

func run(testFileGlobs []string) error {
	flag.Set("test.v", "1")
	testFiles, err := collectTestFiles(testFileGlobs)
	if err != nil {
		log.Fatal(err)
	}

	var allTestCases []pat.TestCase
	for _, testFile := range testFiles {
		prt, err := pat.NewPromRuleTestFromFile(testFile)
		if err != nil {
			log.Fatal(err)
		}

		testCasesForFile, err := prt.GenerateTestCases()
		if err != nil {
			log.Fatal(err)
		}
		for _, tc := range testCasesForFile {
			allTestCases = append(allTestCases, tc)
		}
	}

	if len(allTestCases) == 0 {
		fmt.Println("WARNING: No tests discovered. Exiting...")
		return nil
	}

	testRunner := pat.GoTestRunner{}
	retCode := testRunner.RunTests(allTestCases)
	if retCode != 0 {
		return errors.New("Bad return code from test")
	}
	return nil
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "pat <test-file-globs...>",
	Short: "Prometheus Alert Testing utility",
	Run: func(cmd *cobra.Command, args []string) {
		run(args)
	},
	Args: cobra.MinimumNArgs(1),
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

