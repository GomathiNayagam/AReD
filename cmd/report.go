// Copyright © 2017 Will Rowe <will.rowe@stfc.ac.uk>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/will-rowe/groot/src/misc"
	"github.com/will-rowe/groot/src/reporting"
	"log"
	"os"
	"runtime"
	"strings"
)

// the command line arguments
var (
	bamFile   *string  // a BAM file to generate report from
	covCutoff *float64 // breadth of coverage theshold
)

// the report command (used by cobra)
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report from the output of groot align",
	Long:  `Generate a report from the output of groot align.
	Currently only reports: gene, length, read count`,
	Run: func(cmd *cobra.Command, args []string) {
		runReport()
	},
}

/*
  A function to initalise the command line arguments
*/
func init() {
	RootCmd.AddCommand(reportCmd)
	bamFile = reportCmd.Flags().StringP("bamFile", "i", "", "BAM file generated by groot alignment (will use STDIN if not provided)")
	covCutoff = reportCmd.Flags().Float64P("covCutoff", "c", 0.9, "coverage cutoff for reporting ARGs")
}

/*
  A function to check user supplied parameters
*/
func reportParamCheck() error {
	// if no BAM files provided, check STDIN
	if *bamFile == "" {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return errors.New(fmt.Sprintf("error with STDIN"))
		}
		if (stat.Mode() & os.ModeNamedPipe) == 0 {
			return errors.New(fmt.Sprintf("no STDIN found"))
		}
		log.Printf("\tBAM file: using STDIN")
		// check the provided BAM files
	} else {
		if _, err := os.Stat(*bamFile); err != nil {
			if os.IsNotExist(err) {
				return errors.New(fmt.Sprintf("BAM file does not exist: %v", *bamFile))
			} else {
				return errors.New(fmt.Sprintf("can't access BAM file (check permissions): %v", *bamFile))
			}
		}
		splitFilename := strings.Split(*bamFile, ".")
		if splitFilename[len(splitFilename)-1] != "bam" {
			return errors.New(fmt.Sprintf("the BAM file does not have a `.bam` extension: %v", *bamFile))
		}
		log.Printf("\tBAM file: %v", *bamFile)
	}
	if *covCutoff > 1.0 {
		return errors.New(fmt.Sprintf("supplied coverage cutoff exceeds 1.0 (100%): %v", *covCutoff))
	}
	// set number of processors to use
	if *proc <= 0 || *proc > runtime.NumCPU() {
		*proc = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*proc)
	return nil
}

/*
  The main function for the align sub-command
*/
func runReport() {
	// set up logging
	logFH, err := os.OpenFile("groot-report.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFH.Close()
	log.SetOutput(logFH)
	log.Printf("starting the report command")
	// check the supplied files and then log some stuff
	log.Printf("checking parameters...")
	misc.ErrorCheck(reportParamCheck())
	log.Printf("\tcoverage cutoff: %.2f", *covCutoff)
	log.Printf("\tprocessors: %d", *proc)
	bamReader := reporting.NewBAMreader()
	if *bamFile != "" {
		bamReader.InputFile = *bamFile
	}
	bamReader.CoverageCutoff = *covCutoff
	bamReader.Run()
	log.Println("finished")

	/*

	   load the graph back in - once annotated ARGs, use the clusters to decide most likely annotation?

	*/

} // end of report main function
