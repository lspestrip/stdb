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
	"fmt"
	"path"
	"time"
	"strings"
)

// TestFile holds information about a FITS file containing the data acquired
// during a test. Since FITS files are always converted from some other kind
// of file saved by the machines in the Bicocca labs, the field InputFileName
// points to the source file used to create the FITS file.
type TestFile struct {
	InputFileName string // Name of the file used to produce the FITS file

	FitsChecksum string // Checksum of the FITS file
	CreationDate time.Time // Time  when the acquisition of the data stopped

	TimeSpanSec float32 // Length of the test, in seconds
	NumOfSamples int // Number of samples acquired during the test
}

// FileType returns a string identifiying the type of the file. It is used
// to determine how to read a file containing the data acquired during a
// test.
func FileType(filepath string) (string, error) {
	// At the moment, we just consider the extension. I think that in the
	// future we'll have to open the file and hunt for known patterns:
	// in that case, having an "error" return value will be handy.
	fileExt := path.Ext(filepath)
	switch strings.ToLower(fileExt) {
		case ".xls":
			return "keithley", nil
	}

	return fmt.Sprintf("unknown(\"%s\")", fileExt), nil
}
