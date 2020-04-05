// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// + build example
//
// This build tag means that "go install golang.org/x/exp/shiny/..." doesn't
// install this example program. Use "go run main.go" to run it or "go install
// -tags=example" to install it.

// Basic is a basic example of a graphical application.
package main

import (
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"

	"github.com/janezkenda/chip8/chip8"
)

var keys = map[rune]byte{
	'1': 0x0,
	'2': 0x1,
	'3': 0x2,
	'4': 0x3,
	'q': 0x4,
	'w': 0x5,
	'e': 0x6,
	'r': 0x7,
	'a': 0x8,
	's': 0x9,
	'd': 0xa,
	'f': 0xb,
	'y': 0xc,
	'x': 0xd,
	'c': 0xe,
	'v': 0xf,
}

func main() {
	rand.Seed(time.Now().UnixNano())
	driver.Main(func(s screen.Screen) {
		pr, err := ioutil.ReadFile("roms/MAZE")
		if err != nil {
			log.Fatal(err)
		}

		screenSize := image.Rect(0, 0, 640, 320)

		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title:  "CHIP-8",
			Width:  screenSize.Max.X,
			Height: screenSize.Max.Y,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		c8 := chip8.Init(nil)
		go c8.RunProgram(pr)

		go func() {
			for {
				src := c8.GetFrame(screenSize.Max.X, screenSize.Max.Y)

				b, err := s.NewBuffer(screenSize.Max)
				if err != nil {
					log.Fatal(err)
				}

				draw.Draw(b.RGBA(), b.RGBA().Bounds(), src, image.Point{}, draw.Src)

				w.Upload(image.Point{}, b, b.Bounds())
				w.Publish()
				b.Release()
			}
		}()

		for {
			e := w.NextEvent()

			// This print message is to help programmers learn what events this
			// example program generates. A real program shouldn't print such
			// messages; they're not important to end users.
			/*format := "got %#v\n"
			if _, ok := e.(fmt.Stringer); ok {
				format = "got %v\n"
			}
			fmt.Printf(format, e)*/

			switch e := e.(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if k, ok := keys[e.Rune]; ok {
					go c8.SendKey(k, e.Direction == key.DirPress)
				}

				if e.Code == key.CodeEscape {
					return
				}

			case error:
				log.Print(e)
			}
		}
	})
}
