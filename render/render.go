package render

import (
	"fmt"

	"github.com/Asiandayboy/CLITextEditor/util/ansi"
)

const (
	Reset string = "\033[0m"

	TopLCorner string = "\u250C"
	TopRCorner string = "\u2510"
	BotLCorner string = "\u2514"
	BotRCorner string = "\u2518"
	Horizontal string = "\u2500"
	Vertical   string = "\u2502"

	DoubleTopLCorner string = "\u2554"
	DoubleTopRCorner string = "\u2557"
	DoubleBotLCorner string = "\u255A"
	DoubleBotRCorner string = "\u255D"
	DoubleHorizontal string = "\u2550"
	DoubleVertical   string = "\u2551"
)

type Box struct {
	Width, Height int
	X, Y          int
	BorderColor   ansi.RGBColor // defaults to white if excluded (255, 255, 255)
	FillColor     ansi.RGBColor // exclude this field if you don't want a fill color
	BorderStyle   string        // "double" or "single"; defaults to single
}

/*
This function draws a box on the terminal screen given a width and height,
as well as x and y coordinates to position the box relative to the current
cursor's position.

If the resetCursor flag is set to true, the cursor will be reset to the
top left of the screen before applying the position

The X and Y positions are capped at the current window size of your terminal
the moment this function is called. So make sure you have enough space for the
size of your box to render from the position
*/
func DrawBox(b Box, resetCursor bool) {
	if resetCursor {
		ansi.MoveCursor(0, 0)
	}

	if b.BorderStyle == "" {
		b.BorderStyle = "single"
	}

	var lines []string
	if b.BorderStyle == "single" {
		lines = []string{
			TopLCorner, TopRCorner,
			BotLCorner, BotRCorner,
			Horizontal, Vertical,
		}
	} else if b.BorderStyle == "double" {
		lines = []string{
			DoubleTopLCorner, DoubleTopRCorner,
			DoubleBotLCorner, DoubleBotRCorner,
			DoubleHorizontal, DoubleVertical,
		}
	}

	if b.X > 0 {
		ansi.MoveCursorRight(b.X)
	}
	if b.Y > 0 {
		ansi.MoveCursorDown(b.Y)
	}

	var color string = ansi.NewRGBColor(255, 255, 255).ToFgColorANSI()

	if b.BorderColor != ansi.NO_COLOR {
		color = b.BorderColor.ToFgColorANSI()
	}

	if b.FillColor != ansi.NO_COLOR {
		color = ansi.CombineFgAndBgColorANSI(b.BorderColor, b.FillColor)
	}

	// begin rendering box
	fmt.Print(color + lines[0] + Reset)
	for range b.Width - 2 { // -2 for the left and right borders
		fmt.Print(color + lines[4] + Reset)
	}
	fmt.Println(color + lines[1] + Reset)

	for range b.Height - 2 { // - 2 for the top border and bottom border
		if b.X > 0 {
			ansi.MoveCursorRight(b.X)
		}
		fmt.Print(color + lines[5] + Reset)
		fmt.Printf("%s%*s%s\n", color, b.Width-1, lines[5], Reset)
	}

	if b.X > 0 {
		ansi.MoveCursorRight(b.X)
	}

	fmt.Print(color + lines[2] + Reset)
	for range b.Width - 2 { // -2 for the left and right borders
		fmt.Print(color + lines[4] + Reset)
	}
	fmt.Print(color + lines[3] + Reset)
}

type Line struct {
	Length          int
	X, Y            int
	LineColor       ansi.RGBColor // defaults to white (255, 255, 255)
	BackgroundColor ansi.RGBColor // exclude this if you don't want a background color
	LineStyle       string        // "dashed", "solid", or "double"; defaults to "solid"
}

func DrawVerticalLine(l Line, resetCurosr bool) {
	if resetCurosr {
		ansi.MoveCursor(0, 0)
	}

	var line string = Vertical
	if l.LineStyle == "dashed" {
		line = "|"
	} else if l.LineStyle == "solid" {
		line = Vertical
	} else if l.LineStyle == "double" {
		line = DoubleVertical
	}

	ansi.MoveCursorRight(l.X)
	ansi.MoveCursorDown(l.Y)

	color := ansi.NewRGBColor(255, 255, 255).ToFgColorANSI()

	if l.LineColor != ansi.NO_COLOR {
		color = l.LineColor.ToFgColorANSI()
	}

	if l.BackgroundColor != ansi.NO_COLOR {
		color = ansi.CombineFgAndBgColorANSI(l.LineColor, l.BackgroundColor)
	}

	for range l.Length {
		fmt.Println(color + line + Reset)
		ansi.MoveCursorRight(l.X)
	}

}
