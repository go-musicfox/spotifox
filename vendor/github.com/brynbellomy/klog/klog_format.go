// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Log formatting support

package klog

import (
	"bytes"
)

// Formatter is a pluggable formatter for klog, allowing external users to provider custom formatting.
type Formatter interface {

	// FormatHeader writes a formatted header to ioBuf.
	// This function is presumed to be THREADSAFE.
	FormatHeader(inSeverity string, inFile string, inLine int, ioBuf *bytes.Buffer)
}

// SetFormatter sets the global/default formatter.
//
// By default, the formatter is set to nil, meaning the historcal glog/klog formatter is used.
func SetFormatter(inFormatter Formatter) {
	logging.formatter = inFormatter
}

// FmtConstWidth is a basic formatter that makes reasonable attempts to make the header length a constant width,
// improving readability. It also can insert console color codes so each severity level is a different color.
type FmtConstWidth struct {
	// FileNameCharWidth is the number of chars to use from the given file name.
	// Filenames shorter than this are padded with spaces.
	// If 0, file names are not printed.
	FileNameCharWidth int

	// If set, color codes will be inserted
	UseColor bool
}

// FormatHeader -- see interface Formatter
func (f *FmtConstWidth) FormatHeader(inSeverity string, inFile string, inLine int, buf *bytes.Buffer) {
	var (
		tmp [64]byte
	)

	sevChar := inSeverity[0]
	sz := 0

	usingColor := f.UseColor
	if usingColor {
		var color byte
		switch sevChar {
		case 'W':
			color = yellow
		case 'E', 'F':
			color = lightRed
		case 'S':
			color = green
		case 'D':
			color = magenta
		default:
			color = dim
		}
		sz += AppendColorCode(color, tmp[sz:])
	}

	tmp[sz] = sevChar
	sz++

	sz += AppendTimestamp(tmp[sz:])
	tmp[sz] = ' '
	sz++
	buf.Write(tmp[:sz])
	sz = 0

	if segSz := f.FileNameCharWidth; segSz > 0 {
		strLen := len(inFile)
		padLen := segSz - strLen
		if padLen < 0 {
			buf.Write([]byte(inFile))
		} else {
			for ; sz < padLen; sz++ {
				tmp[sz] = ' '
			}
			for ; sz < segSz; sz++ {
				tmp[sz] = inFile[sz-padLen]
			}
		}
		tmp[sz] = ':'
		sz++
		if inLine < 10000 {
			sz += AppendNDigits(4, inLine, tmp[sz:], '0')
		} else {
			sz += AppendDigits(inLine, tmp[sz:])
		}
	}
	tmp[sz] = ']'
	tmp[sz+1] = ' '
	sz += 2

	if usingColor {
		sz += AppendColorCode(byte(noColor), tmp[sz:])
	}

	buf.Write(tmp[:sz])
}

// AppendTimestamp appends a glog/klog-style timestamp to the given slice,
// returning how many bytes were written.
//
// Pre: len(buf) >= 20
func AppendTimestamp(buf []byte) int {

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	now := timeNow()
	_, month, day := now.Date()
	hour, min, sec := now.Clock()

	// mmdd hh:mm:ss.uuuuuu
	sz := 0
	buf[sz+0] = '0' + byte(month)/10
	buf[sz+1] = '0' + byte(month)%10
	buf[sz+2] = '0' + byte(day/10)
	buf[sz+3] = '0' + byte(day%10)
	buf[sz+4] = ' '
	sz += 5
	buf[sz+0] = '0' + byte(hour/10)
	buf[sz+1] = '0' + byte(hour%10)
	buf[sz+2] = ':'
	buf[sz+3] = '0' + byte(min/10)
	buf[sz+4] = '0' + byte(min%10)
	buf[sz+5] = ':'
	buf[sz+6] = '0' + byte(sec/10)
	buf[sz+7] = '0' + byte(sec%10)
	buf[sz+8] = '.'
	sz += 9
	sz += AppendNDigits(6, now.Nanosecond()/1000, buf[sz:], '0')

	return sz
}

// AppendDigits appends the base 10 value to the given buffer, returning the number of bytes written.
//
// Pre: inValue > 0
func AppendDigits(inValue int, buf []byte) int {
	sz := 0
	for ; inValue > 0; sz++ {
		buf[sz] = '0' + byte(inValue%10)
		inValue /= 10
	}
	// Reverse the digits in place
	for i := sz/2 - 1; i >= 0; i-- {
		idx := sz - 1 - i
		tmp := buf[i]
		buf[i] = buf[idx]
		buf[idx] = tmp
	}
	return sz
}

// AppendNDigits formats an n-digit integer to the given buffer, padding as needed,
// returning the number of bytes written.
//
// Pre: len(buf) >= inNumDigits
func AppendNDigits(inNumDigits int, inValue int, buf []byte, inPad byte) int {
	j := inNumDigits - 1
	for ; j >= 0 && inValue > 0; j-- {
		buf[j] = '0' + byte(inValue%10)
		inValue /= 10
	}
	for ; j >= 0; j-- {
		buf[j] = inPad
	}
	return inNumDigits
}

// AppendColorCode appends the console color code to the given slice,
// returning how many bytes were appended.
func AppendColorCode(inColorCode byte, buf []byte) int {
	buf[0] = '\x1b'
	buf[1] = '['
	sz := 2

	code := inColorCode
	if code >= 100 {
		digit := '0' + code/100
		buf[sz] = digit
		sz++
		code -= digit * 100
	}
	if code >= 10 {
		digit := inColorCode / 10
		buf[sz] = '0' + digit
		sz++
		code -= digit * 10
	}
	buf[sz] = '0' + code

	buf[sz+1] = 'm'
	sz += 2

	return sz
}

const (
	noColor      = 0
	bold         = 1
	dim          = 2
	underline    = 3
	invert       = 7
	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGrey     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
)
