package ansi

import "fmt"

var NO_COLOR RGBColor = RGBColor{}

/*
Represents an RGB color that can be used to style your text
Use NewRGBColor to create a new RGBColor
*/
type RGBColor struct{ R, G, B uint8 }

/*
Returns the ANSI escape code representation of the foreground RGB color
*/
func (c RGBColor) ToFgColorANSI() string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", c.R, c.G, c.B)
}

/*
Returns the ANSI escape code representation of the background RGB color
*/
func (c RGBColor) ToBgColorANSI() string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", c.R, c.G, c.B)
}

/*
Returns an RGBColor that can be used to style your text
*/
func NewRGBColor(R, G, B uint8) RGBColor {
	return RGBColor{R, G, B}
}

/*
Returns the ANSI escape code repreentation of the
foreground and background RGB color as a single string
*/
func CombineFgAndBgColorANSI(fgColor RGBColor, bgColor RGBColor) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm",
		fgColor.R, fgColor.G, fgColor.B,
		bgColor.R, bgColor.G, bgColor.B,
	)
}
