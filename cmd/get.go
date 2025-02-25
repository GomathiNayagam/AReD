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
	"archive/tar"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/will-rowe/AReD/src/version"
)

// available databases to download
var availDb = []string{"arg-annot", "resfinder", "card", "AReD-db", "AReD-core-db"}
var availIdent = []string{"90"}
var md5sums = map[string]string{
	"arg-annot.90":     "d5398b7bd40d7e872c3e4a689cee4726",
	"resfinder.90":     "de34ab790693cb7c7b656d537ec40f05",
	"card.90":          "23b24d37edfd20016c2d8b5a522a4d10",
	"AReD-db.90":       "2cbbe9a89c2ce23c09575198832250d3",
	"AReD-core-db.90":  "f3cac49ff44624a26ea2d92171a73174",
}

// dbURL to download databases from
var dbURL = fmt.Sprintf("https://github.com/will-rowe/AReD/raw/master/db/clustered-ARG-databases/%v/", version.GetBaseVersion())

// the command line arguments
var (
	database *string // the database to download
	identity *string // the sequence identity used to cluster the database
	dbDir    *string // the location to store the database
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Download a pre-clustered ARG database",
	Long:  `Download a pre-clustered ARG database`,
	Run: func(cmd *cobra.Command, args []string) {
		runGet()
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
	database = getCmd.Flags().StringP("database", "d", "arg-annot", "database to download (please choose: arg-annot/resfinder/card/AReD-db/AReD-core-db)")
	identity = getCmd.Flags().String("identity", "90", "the sequence identity used to cluster the database (only 90 available atm)")
	dbDir = getCmd.PersistentFlags().StringP("out", "o", ".", "directory to save the database to")
}

/*
  A function to check user supplied parameters
*/
func getParamCheck() error {
	// check requested db exists in AReD records
	checkPass := false
	for _, avail := range availDb {
		if *database == avail {
			checkPass = true
		}
	}
	if checkPass == false {
		return fmt.Errorf("unrecognised DB: %v\n\nplease choose either: arg-annot/resfinder/card/AReD-db/AReD-core-db", *database)
	}
	checkPass = false
	for _, avail := range availIdent {
		if *identity == avail {
			checkPass = true
		}
	}
	if checkPass == false {
		return fmt.Errorf("identity value not available: %v\n\nplease choose either: 90", *identity)
	}
	// setup the dbDir
	if _, err := os.Stat(*dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(*dbDir, 0700); err != nil {
			return fmt.Errorf("directory creation failed: %v\n\ncan't create specified output directory for the database", *dbDir)
		}
	}
	return nil
}

/*
  A function to download the database tarball
*/
func DownloadFile(savePath string, url string) error {
	outFile, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		return err
	}
	return nil
}

/*
  A function to calculate md5
*/
func getMD5(savePath string) error {
	var dbMD5 string
	file, err := os.Open(savePath)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	hashInBytes := hash.Sum(nil)[:16]
	dbMD5 = hex.EncodeToString(hashInBytes)
	lookup := fmt.Sprintf("%v.%v", *database, *identity)
	if dbMD5 != md5sums[lookup] {
		return errors.New("md5sum for downloaded tarball did not match record")
	}
	return nil
}

/*
  The main function for the get sub-command
*/
func runGet() {
	if err := getParamCheck(); err != nil {
		fmt.Println("could not run AReD get...")
		fmt.Println(err)
		os.Exit(1)
	}
	// download the db
	fmt.Printf("downloading the pre-clustered %v database...\n", *database)
	dbName := fmt.Sprintf("%v.%v", *database, *identity)
	dbURL += dbName
	dbURL += ".tar"
	if err := DownloadFile("tmp.tar", dbURL); err != nil {
		fmt.Println("could not download the tarball")
		fmt.Println(err)
		os.Exit(1)
	}
	// unpack the db
	fmt.Println("unpacking...")
	if err := getMD5("tmp.tar"); err != nil {
		fmt.Println("could not unpack the tarball")
		fmt.Println(err)
		os.Exit(1)
	}
	fh, err := os.Open("tmp.tar")
	if err != nil {
		fmt.Println("could not unpack the tarball")
		fmt.Println(err)
		os.Exit(1)
	}
	defer fh.Close()

	if err := Untar("tmp", fh); err != nil {
		fmt.Println("could not unpack the tarball")
		fmt.Println(err)
		os.Exit(1)
	}
	tmpDb := fmt.Sprintf("tmp/%v", dbName)
	dbSave := fmt.Sprintf("%v/%v.%v", *dbDir, *database, *identity)
	if err := os.Rename(tmpDb, dbSave); err != nil {
		fmt.Println("could not save db to specified directory")
		os.Exit(1)
	}
	// finished
	if err := os.Remove("tmp.tar"); err != nil {
		fmt.Println("could not cleanup...")
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.Remove("tmp"); err != nil {
		fmt.Println("could not cleanup...")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("database saved to: %v\n", dbSave)
	fmt.Printf("now run `AReD index -m %v` or `AReD index --help` for full options\n", dbSave)
}

// Untar will untar an archive
func Untar(dst string, fileReader io.Reader) error {
	if err := os.MkdirAll(dst, os.FileMode(0755)); err != nil {
		return err
	}

	tarBallReader := tar.NewReader(fileReader)
	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		filename := fmt.Sprintf("%v/%v", dst, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filename, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			writer, err := os.Create(filename)
			if err != nil {
				return err
			}
			io.Copy(writer, tarBallReader)
			if err := os.Chmod(filename, os.FileMode(header.Mode)); err != nil {
				return err
			}
			writer.Close()
		default:
			return fmt.Errorf("unable to untar type : %c in file %s", header.Typeflag, filename)
		}
	}
	return nil
}

