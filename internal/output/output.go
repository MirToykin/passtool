package output

import (
	"github.com/fatih/color"
	"os"
)

// colorWrapper wrapper function that returns function for printing with the given color
func colorWrapper(colorAttr color.Attribute) func(msg string, a ...interface{}) {
	return func(msg string, a ...interface{}) {
		color.New(colorAttr).PrintfFunc()(msg, a...)
	}
}

type Out struct{}

// Simple for printing regular text
func (o Out) Simple(msg string, a ...interface{}) {
	colorWrapper(color.FgWhite)(msg, a...)
}

// Simpleln for printing regular text with new line
func (o Out) Simpleln(msg string, a ...interface{}) {
	o.Simple(msg+"\n", a...)
}

// Info for printing info text
func (o Out) Info(msg string, a ...interface{}) {
	colorWrapper(color.FgBlue)(msg, a...)
}

// Infoln for printing info text with new line
func (o Out) Infoln(msg string, a ...interface{}) {
	o.Info(msg+"\n", a...)
}

// Header for printing header text
func (o Out) Header(msg string, a ...interface{}) {
	colorWrapper(color.FgMagenta)(msg+"\n", a...)
}

// Success for printing success message
func (o Out) Success(msg string, a ...interface{}) {
	colorWrapper(color.FgGreen)(msg+"\n", a...)
}

// Warning for printing warning message
func (o Out) Warning(msg string, a ...interface{}) {
	colorWrapper(color.FgYellow)(msg+"\n", a...)
}

// Error for printing error message
func (o Out) Error(msg string, a ...interface{}) {
	colorWrapper(color.FgRed)(msg, a...)
}

// ErrorWithExit for printing error message with the following stopping execution
func (o Out) ErrorWithExit(msg string, a ...interface{}) {
	o.Error(msg+"\n", a...)
	os.Exit(1)
}

// New returns new instance of Out
func New() Out {
	return Out{}
}
