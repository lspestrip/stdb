// Copyright © 2017 Maurizio Tomasi <maurizio.tomasi@unimi.it>
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

package convert

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/astrogo/fitsio"
)

func TestExecutionTimeConversion(t *testing.T) {
	val, err := parseExecutionTime("01:02:03:04")
	if err != nil {
		t.Errorf("unable to parse the execution time: %v", err)
	} else if int64(val) != 86400 + 2 * 3600 + 3 * 60 + 4 {
		t.Errorf("wrong execution time: %f", val)
	}
}

// gunzipFile decompresses a .gz file into "out". This function is
// necessary because astrogo/fitsio does not support reading from
// gzipped streams (despite the fact that it does not seem to have
// problems in writing into them)
func gunzipFile(sourcePath string, out io.Writer) error {
	f, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer fz.Close()

	if _, err := io.Copy(out, fz); err != nil {
		return err
	}

	return nil
}

type fitsContents struct {
	headerCards map[string]interface{}
	table dataTable
}

func readFitsFileData(filePath string) (fitsContents, error) {
	var result fitsContents

	// Since the input FITS file is gzipped, create a new temporary
	// file that will containg the uncompressed file
	f, err := ioutil.TempFile("", "gunzipped_fits")
	if err != nil {
		return result, err
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// Uncompress
	if err := gunzipFile(filePath, f); err != nil {
		return result, err
	}

	// Move the pointer to the beginning of the file
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return result, err
	}

	fits, err := fitsio.Open(f)
	if err != nil {
		return result, err
	}
	defer fits.Close()

	dataHDU, ok := fits.HDU(1).(*fitsio.Table)
	if !ok {
		return result, fmt.Errorf("no table HDU found in the FITS file")
	}
	defer dataHDU.Close()
	numOfRows := dataHDU.NumRows()
	
	hdr := dataHDU.Header()
	result.headerCards = make(map[string]interface{}, 0)
	for _, k := range hdr.Keys() {
		result.headerCards[k] = hdr.Get(k).Value
	}

	result.table.Columns = make([]dataColumn, dataHDU.NumCols())
	for i, curCol := range dataHDU.Cols() {
		result.table.Columns[i].name = curCol.Name
		result.table.Columns[i].values = make([]float64, numOfRows)
	}

	rows, err := dataHDU.Read(0, numOfRows)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	// Make room for the elements of one row
	elements := make([]interface{}, len(result.table.Columns))
	for i, curCol := range dataHDU.Cols() {
		elements[i] = reflect.New(curCol.Type()).Interface()
	}

	for curRow := 0; rows.Next(); curRow++ {
		if err := rows.Scan(elements...); err != nil {
			return result, err
		}
		for curCol := range dataHDU.Cols() {
			result.table.Columns[curCol].values[curRow] = float64(*elements[curCol].(*float32))
		}
	}

	return result, err
}

func areCloseEnough(a float64, b float64) bool {
	diff := math.Abs(a - b)
	if diff == 0.0 {
		return true
	}

	return diff / math.Abs(a) < 1.0e-6
}

func TestKeithleyConversion(t *testing.T) {
	sourceFilePath := path.Join("..", "testdata", "keithley_file.xls")
	destFile, err := ioutil.TempFile("", "stdb_convert_keithley.fits.gz")
	if err != nil {
		t.Logf("unable to create output file: %v", err)
		t.FailNow()
	}
	defer os.Remove(destFile.Name())

	// Create a new FITS file
	testFile, err := KeithleyXlsToFits(sourceFilePath, destFile, []fitsio.Card {
		{ Name: "test", Value: 123 },
	})
	if err != nil {
		t.Logf("unable to create output file \"%s\": %v", destFile.Name(), err)
	}
	destFile.Close()

	// Check that the FITS file contains valid data
	dataFromFits, err := readFitsFileData(destFile.Name())
	if err != nil {
		t.Logf("unable to read FITS file \"%s\": %v", destFile.Name(), err)
		t.FailNow()
	}

	// Verify that the metadata are correct
	refMetadata := map[string]interface{} {
		"testname": "If_vs_Vf_Det3#1@1",
		"mode": "Sweeping",
		"speed": "Normal",
		"swdelay": 0.0,
        "coord": "0,0",
        "acqtime": "2017-05-18T10:38:25Z",
        "clarver": "V1.1",
        "extime": 5.0,
        "interlck": "High Voltage Enabled",
		"test": 123,
	}
	for key, refVal := range refMetadata {
		curVal, ok := dataFromFits.headerCards[key]
		if ! ok {
			t.Errorf("missing card \"%s\" from FITS header", key)
		} else {
			if curVal != refVal {
				t.Errorf("wrong card in FITS header: %v != %v", curVal, refVal)
			}
		}
	}

	// Verify that the data table is correct
	if len(dataFromFits.table.Columns) != 4 {
		t.Errorf("%d columns found in FITS file, instead of 4", len(dataFromFits.table.Columns))
	}

	for idx, refName := range map[int]string{ 0: "EmitterI", 1: "EmitterV", 2: "BaseI", 3: "BaseV" } {
		if colName := dataFromFits.table.Columns[idx].name; colName != refName {
			t.Errorf("column 0 has a wrong name (\"%s\" != \"%s\")", colName, refName)
		}
	}

	for idx, refVal := range []float64{ 7.409528e-12, 0.5, 0.0, -8.399948e-2 } {
		if curVal := dataFromFits.table.Columns[idx].values[0]; ! areCloseEnough(curVal, refVal) {
			t.Errorf("wrong value in the first line: %e != %e (column %d)", curVal, refVal, idx)
		}
	}

	for idx, refVal := range []float64{ -9.999024e-5, 0.5, 1e-4, 0.7755449 } {
		lastIdx := len(dataFromFits.table.Columns[idx].values) - 1
		if curVal := dataFromFits.table.Columns[idx].values[lastIdx]; ! areCloseEnough(curVal, refVal) {
			t.Errorf("wrong value in the last line: %e != %e (column %d)", curVal, refVal, idx)
		}
	}

	// Check that testFile contains correct information
	if testFile.InputFileName != sourceFilePath {
		t.Errorf("Wrong InputFileName field: \"%s\"", testFile.InputFileName)
	}
}