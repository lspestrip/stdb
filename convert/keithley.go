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
	"math"
	"strconv"
	"strings"

	"github.com/extrame/xls"
	"github.com/astrogo/fitsio"
	"time"
)

const (
	expectedNumOfSheets = 3
)

type dataColumn struct {
	name string
	values []float64
}

type dataTable struct {
	Columns []dataColumn
}

type metadata struct {
	TestName string
	Mode string
	Speed string
	SweepDelay float64
	HoldTime float64
	SiteCoordinate string
	LastExecuted time.Time
	ClariusVersion string
	ExecutionTimeSec float64
	Interlock string
}

func fillColumnData(sheet *xls.WorkSheet, colNum int, data []float64) {
	numOfRows := int(sheet.MaxRow)
	for i := 1; i <= numOfRows; i++ {
		value, err := strconv.ParseFloat(sheet.Row(i).Col(colNum), 64)
		if err != nil {
			value = math.NaN()
		}

		data[i - 1] = value
	}
}

// importDataTable loads the data from the first worksheet in the Keithley
// Excel file (the one named "RunNN"). This contains all the voltage/current
// values measured during the test
func importDataTable(sheet *xls.WorkSheet, table *dataTable) error {
	if sheet.MaxRow == 0 {
		// This sheet is empty
		return nil
	}

	table.Columns = []dataColumn{}
	// Determine how many columns are in this worksheet
	headerRow := sheet.Row(0)
	for curColNum := headerRow.FirstCol(); curColNum <= headerRow.LastCol(); curColNum++ {
		curHeader := headerRow.Col(curColNum)
		if curHeader != "" {
			curData := make([]float64, sheet.MaxRow)
			fillColumnData(sheet, curColNum, curData)
			curColumn := dataColumn{ name: curHeader, values: curData }
			table.Columns = append(table.Columns, curColumn)
		}
	}

	return nil
}

// parseExecution time returns the number of seconds equivalent to the time in "s"
// The string "s" must be in the form
//
//     DD:HH:MM:SS
//
// where DD is the number of days, HH the number of hours,
// MM the number of minutes, and SS the number of seconds
func parseExecutionTime(s string) (float64, error) {
	elements := strings.Split(s, ":")
	if len(elements) != 4 {
		return 0.0, fmt.Errorf("wrong execution time \"%s\"", s)
	}
	var result float64

	factor := []float64{ 86400.0, 3600.0, 60.0, 1.0 }
	for idx, curElem := range elements {
		curVal, err := strconv.Atoi(curElem)
		if err != nil {
			return 0.0, fmt.Errorf("wrong execution time \"%s\" near \"%s\"", s, curElem)
		}
		result += float64(curVal) * factor[idx]
	}

	return result, nil
}

// importMetadata loads the text in the "Settings" worksheet. It contains
// information about the kind of test.
func importMetadata(sheet *xls.WorkSheet, meta *metadata) error {
	if sheet.MaxRow == 0 {
		// This sheet is empty
		return nil
	}

	for curRowNum := 1; curRowNum <= 10; curRowNum++ {
		curKey := sheet.Row(curRowNum).Col(0)
		curValue := sheet.Row(curRowNum).Col(1)

		var err error
		switch strings.ToLower(curKey) {
			case "test name":
				meta.TestName = curValue
			case "mode":
				meta.Mode = curValue
			case "speed":
				meta.Speed = curValue
			case "sweep delay":
				meta.SweepDelay, err = strconv.ParseFloat(curValue, 64)
			case "hold time":
				meta.HoldTime, err = strconv.ParseFloat(curValue, 64)
			case "site coordinate":
				meta.SiteCoordinate = curValue
			case "last executed":
				meta.LastExecuted, err = time.Parse("01/02/2006 15:04:05", curValue)
			case "clarius+ version":
				meta.ClariusVersion = curValue
			case "execution time":
				meta.ExecutionTimeSec, err = parseExecutionTime(curValue)
			case "interlock":
				meta.Interlock = curValue
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func fillFitsTableHeader(fitsTable *fitsio.Table, meta *metadata, fitshdr []fitsio.Card) error {
	hdr := fitsTable.Header()
	for _, card := range fitshdr {
		hdr.Set(card.Name, card.Value, card.Comment)
	}
	return hdr.Append(fitsio.Card{ Name: "testname", Value: meta.TestName, Comment: "Name of the test"},
	                  fitsio.Card{ Name: "mode", Value: meta.Mode },
			          fitsio.Card{ Name: "speed", Value: meta.Speed },
			          fitsio.Card{ Name: "swdelay", Value: meta.SweepDelay, Comment: "Sweep delay" },
			          fitsio.Card{ Name: "coord", Value: meta.SiteCoordinate, Comment: "Site coordinates" },
			          fitsio.Card{ Name: "acqtime", Value: meta.LastExecuted.Format(time.RFC3339), Comment: "Last executed" },
			          fitsio.Card{ Name: "clarver", Value: meta.ClariusVersion, Comment: "Clarius+ version" },
			          fitsio.Card{ Name: "extime", Value: meta.ExecutionTimeSec, Comment: "Execution time [s]" },
			          fitsio.Card{ Name: "interlck", Value: meta.Interlock, Comment: "Interlock" })
}

// saveFitsFile writes the metadata and the data table into a FITS file
func saveGzipFitsFile(table *dataTable,
                      meta *metadata,
					  fitshdr []fitsio.Card,
					  destWriter io.Writer) error {
	zw := gzip.NewWriter(destWriter)
	defer zw.Close()

	f, err := fitsio.Create(zw)
	if err != nil {
		return err
	}
	defer f.Close()

	phdu, err := fitsio.NewPrimaryHDU(nil)
	if err != nil {
		return err
	}
	defer phdu.Close()

	err = f.Write(phdu)
	if err != nil {
		return err
	}

	// Create the binary table
	var columns = make([]fitsio.Column, len(table.Columns))
	for colIdx, dataCol := range table.Columns {
		var unit string
		if strings.HasSuffix(dataCol.name, "I") {
			unit = "A"
		} else if strings.HasSuffix(dataCol.name, "V") {
			unit = "V"
		}
		columns[colIdx] = fitsio.Column{ Name: dataCol.name, Format: "E", Unit: unit, Bscale: 1.0 }
	}
	fitsTable, err := fitsio.NewTable("data", columns, fitsio.BINARY_TBL)
	if err != nil {
		return err
	}
	defer fitsTable.Close()

	if err := fillFitsTableHeader(fitsTable, meta, fitshdr); err != nil {
		return err
	}

	// Copy the value of each sample in the table
	var rowValues = make([]interface{}, len(table.Columns))
	for rowIdx := 0; rowIdx < len(table.Columns[0].values); rowIdx++ {
		for colIdx := 0; colIdx < len(table.Columns); colIdx++ {
			rowValues[colIdx] = float32(table.Columns[colIdx].values[rowIdx])
		}
		if err := fitsTable.Write(rowValues...); err != nil {
			return fmt.Errorf("unable to write %v: %v", rowValues, err)
		}
	}

	return f.Write(fitsTable)
}

// KeithleyXlsToFits converts a XLS file produced by the Keithley acquisition
// machine into a FITS file ready to be copied inside the database
func KeithleyXlsToFits(inputpath string,
                       w io.Writer,
					   fitshdr []fitsio.Card) (TestFile, error) {
	var result TestFile

	xlsFile, err := xls.Open(inputpath, "utf-8")
	if err != nil {
		return result, err
	}

	if numSheets := xlsFile.NumSheets(); numSheets != expectedNumOfSheets {
		return result, fmt.Errorf("unrecognized Keithley Xls format (%d sheets instead of %d)",
		                          numSheets, expectedNumOfSheets)
	}

	var table dataTable
	var meta metadata
	for curSheetIdx := 0; curSheetIdx < xlsFile.NumSheets(); curSheetIdx++ {
		curSheet := xlsFile.GetSheet(curSheetIdx)
		switch curSheetIdx {
			case 0: err = importDataTable(curSheet, &table)
			case 2: err = importMetadata(curSheet, &meta)
		}
		if err != nil {
			return result, err
		}
	}

	if err := saveGzipFitsFile(&table, &meta, fitshdr, w); err != nil {
		return result, err
	}

	result.InputFileName = inputpath
	if len(table.Columns) > 0 {
		result.NumOfSamples = len(table.Columns[0].values)
	} else {
		result.NumOfSamples = 0
	}
	result.CreationDate = meta.LastExecuted
	result.TimeSpanSec = float32(meta.ExecutionTimeSec)

	return result, nil
}