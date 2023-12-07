package output

import (
	"github.com/fatih/color"
	"os"
)

func colorWrapper(colorAttr color.Attribute) func(msg string, a ...interface{}) {
	return func(msg string, a ...interface{}) {
		color.New(colorAttr).PrintfFunc()(msg, a...)
	}
}

type Out struct{}

func (o Out) Info(msg string, a ...interface{}) {
	colorWrapper(color.FgBlue)(msg, a...)
}

func (o Out) Infoln(msg string, a ...interface{}) {
	o.Info(msg+"\n", a...)
}

func (o Out) Header(msg string, a ...interface{}) {
	colorWrapper(color.FgMagenta)(msg+"\n", a...)
}

func (o Out) Success(msg string, a ...interface{}) {
	colorWrapper(color.FgGreen)(msg+"\n", a...)
}

func (o Out) Warning(msg string, a ...interface{}) {
	colorWrapper(color.FgYellow)(msg+"\n", a...)
}

func (o Out) Error(msg string, a ...interface{}) {
	colorWrapper(color.FgRed)(msg, a...)
}

func (o Out) ErrorWithExit(msg string, a ...interface{}) {
	o.Error(msg+"\n", a...)
	os.Exit(1)
}

func New() Out {
	return Out{}
}
