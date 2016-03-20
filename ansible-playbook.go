package main

/*
 * Addresses:
 * Can be used in conjunction with below fix to start ansible with extra-vars parameter without falling over.
 * https://github.com/mitchellh/vagrant/issues/6726
 *
 * References:
 * http://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
 * http://www.darrencoxall.com/golang/executing-commands-in-go/
 * https://groups.google.com/forum/#!msg/golang-nuts/cnG-N3KcoUU/4vR5MIuFDjYJ
 *
 */

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

	var wrapperExecutableName = filepath.Base(os.Args[0])
	var extension = filepath.Ext(wrapperExecutableName)
	var executable = wrapperExecutableName[0 : len(wrapperExecutableName)-len(extension)]
	//	pwd, err := os.Getwd()

	args := make([]string, len(os.Args[1:])+1, len(os.Args[1:])+1)
	args[0] = "/usr/bin/" + executable
	for i, a := range os.Args[1:] {
		// fmt.Printf("arg[%d] is: %s\n", i+1, a)
		args[i+1] = a
	}
	cmd := exec.Command("c:\\cygwin\\bin\\bash.exe", "-c", strings.Join(args[:], " "))

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating "+executable+" wrapper stdout pipeline", err)
		os.Exit(1)
	}

	// echo stdin back to console
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting "+executable+" wrapper", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for "+executable+" wrapper to finish", err)
		os.Exit(1)
	}

	// all good can finish now
	os.Exit(0)

}