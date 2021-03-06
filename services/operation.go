package services

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/fatih/color"
	"github.com/yext/edward/common"
)

type OperationTracker interface {
	Start()
	Success()
	SoftFail(err error)
	Fail(err error)
}

var _ OperationTracker = &CommandTracker{}

// CommandTracker follows an operation executed by running a shell command
type CommandTracker struct {
	Name       string
	OutputFile string
	Logger     common.Logger
	sigChan    chan os.Signal
	endChan    chan struct{}
}

func (c *CommandTracker) printf(format string, v ...interface{}) {
	if c.Logger == nil {
		return
	}
	c.Logger.Printf(format, v...)
}

func (c *CommandTracker) waitForInterrupt() {
	c.sigChan = make(chan os.Signal, 1)
	c.endChan = make(chan struct{}, 1)
	signal.Notify(c.sigChan, os.Interrupt)
	go func() {
		select {
		case _ = <-c.sigChan:
			printResult("Interrupted", color.FgRed)
			c.printf("%v Interrupted\n", c.Name)
			if len(c.OutputFile) > 0 {
				c.printFile(c.OutputFile)
			}
			os.Exit(1)
		case _ = <-c.endChan:
			break
		}
	}()
}

func (c *CommandTracker) endWait() {
	c.endChan <- struct{}{}
}

func (c *CommandTracker) Start() {
	fmt.Printf("%-50s", c.Name+"...")
	c.printf("%v\n", c.Name)
	c.waitForInterrupt()
}

func (c *CommandTracker) Success() {
	printResult("OK", color.FgGreen)
	c.printf("%v Succeeded\n", c.Name)
	c.endWait()
}

func (c *CommandTracker) SoftFail(err error) {
	printResult(err.Error(), color.FgYellow)
	c.printf("%v: %v\n", c.Name, err.Error())
	c.endWait()
}

func (c *CommandTracker) Fail(err error) {
	printResult("Failed", color.FgRed)
	c.printf("%v Failed: %v\n", c.Name, err.Error())
	if len(c.OutputFile) > 0 {
		c.printFile(c.OutputFile)
	}
	c.endWait()
}

func printResult(message string, c color.Attribute) {
	print("[")
	color.Set(c)
	print(message)
	color.Unset()
	println("]")
}

func (c *CommandTracker) printFile(path string) {
	dat, errRead := ioutil.ReadFile(path)
	if errRead != nil {
		c.printf("%v: Error reading operation log (%v)\n", c.Name, errRead)
		log.Println(errRead)
		return
	}
	c.printf("%v", string(dat))
	fmt.Print(string(dat))
}
