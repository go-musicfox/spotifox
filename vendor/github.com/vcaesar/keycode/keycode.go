// Copyright 2016 The go-vgo Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

package keycode

// UMap type map[string]uint16
type UMap map[string]uint16

// MouseMap robotgo hook mouse's code map
var MouseMap = UMap{
	"left":       1,
	"right":      2,
	"center":     3,
	"wheelDown":  4,
	"wheelUp":    5,
	"wheelLeft":  6,
	"wheelRight": 7,
}

// Keycode robotgo hook key's code map
var Keycode = UMap{
	"`": 41,
	"1": 2,
	"2": 3,
	"3": 4,
	"4": 5,
	"5": 6,
	"6": 7,
	"7": 8,
	"8": 9,
	"9": 10,
	"0": 11,
	"-": 12,
	"=": 13,
	//
	"_": 12,
	"+": 13,
	//
	"q":  16,
	"w":  17,
	"e":  18,
	"r":  19,
	"t":  20,
	"y":  21,
	"u":  22,
	"i":  23,
	"o":  24,
	"p":  25,
	"[":  26,
	"]":  27,
	"\\": 43,
	//
	"{": 26,
	"}": 27,
	"|": 43,
	//
	"a": 30,
	"s": 31,
	"d": 32,
	"f": 33,
	"g": 34,
	"h": 35,
	"j": 36,
	"k": 37,
	"l": 38,
	";": 39,
	"'": 40,
	//
	":": 39,
	`"`: 40,
	//
	"z": 44,
	"x": 45,
	"c": 46,
	"v": 47,
	"b": 48,
	"n": 49,
	"m": 50,
	",": 51,
	".": 52,
	"/": 53,
	//
	"<": 51,
	">": 52,
	"?": 53,
	//
	"f1":  59,
	"f2":  60,
	"f3":  61,
	"f4":  62,
	"f5":  63,
	"f6":  64,
	"f7":  65,
	"f8":  66,
	"f9":  67,
	"f10": 68,
	"f11": 69,
	"f12": 70,
	// numeric keypad
	"num1":         79,
	"num2":         80,
	"num3":         81,
	"num4":         75,
	"num5":         76,
	"num6":         77,
	"num7":         71,
	"num8":         72,
	"num9":         73,
	"num0":         82,
	"num_minus":    74,
	"num_plus":     78,
	"num_asterisk": 55,
	"num_slash":    3637,
	"num_enter":    3612,
	// more
	"esc":     1,
	"delete":  14,
	"tab":     15,
	"enter":   28,
	"ctrl":    29,
	"control": 29,
	"shift":   42,
	"rshift":  54,
	"space":   57,
	//
	"alt":     56,
	"ralt":    3640,
	"cmd":     3675,
	"command": 3675,
	"rcmd":    3676,
	//
	"up":    57416,
	"down":  57424,
	"left":  57419,
	"right": 57421,
}

// Special is the special key map
var Special = map[string]string{
	"~": "`",
	"!": "1",
	"@": "2",
	"#": "3",
	"$": "4",
	"%": "5",
	"^": "6",
	"&": "7",
	"*": "8",
	"(": "9",
	")": "0",
	"_": "-",
	"+": "=",
	"{": "[",
	"}": "]",
	"|": "\\",
	":": ";",
	`"`: "'",
	"<": ",",
	">": ".",
	"?": "/",
}
