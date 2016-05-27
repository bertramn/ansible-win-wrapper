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
	"github.com/vaughan0/go-ini"
	"os"
	"os/exec"
	"strings"
	"path/filepath"
)

func main() {

	var wrapperExecutableName = filepath.Base(os.Args[0])
	var wrapperExecutableDir = filepath.Dir(os.Args[0])
	var extension = filepath.Ext(wrapperExecutableName)
	var executable = wrapperExecutableName[0 : len(wrapperExecutableName) - len(extension)]
	// default cygwin home
	var cygHome = "c:/cygwin"
	// now check if we need to override the path with env variable
	var cygHomeEnv = os.Getenv("CYGWIN_HOME")
	if cygHomeEnv != "" {
		cygHome = cygHomeEnv
	} else {
		// try to lookup from ini file
		var iniFile = filepath.ToSlash(wrapperExecutableDir + "/cygwin.ini")
		file, err := ini.LoadFile(iniFile)
		if err == nil {
			p, ok := file.Get("cygwin", "home")
			if ok {
				cygHome = p
			}
		}
	}

	args := make([]string, len(os.Args[1:]) + 1, len(os.Args[1:]) + 1)
	args[0] = "/usr/bin/" + executable
	for i, a := range os.Args[1:] {
		// fmt.Printf("arg[%d] is: %s\n", i+1, a)
		args[i + 1] = a
	}

	var bashCmd = filepath.ToSlash(cygHome + "/bin/bash.exe")
	//fmt.Println("execute: " + bashCmd)
	cmd := exec.Command(bashCmd, "-c", strings.Join(args[:], " "))

	// echo stdin back to console
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating " + executable + " wrapper stdout pipeline", err)
		os.Exit(1)
	}

	outScanner := bufio.NewScanner(cmdReader)
	go func() {
		for outScanner.Scan() {
			fmt.Printf("%s\n", outScanner.Text())
		}
	}()

	// echo stderr back to console
	errReader, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating " + executable + " wrapper stderr pipeline", err)
		os.Exit(1)
	}

	errScanner := bufio.NewScanner(errReader)
	go func() {
		for errScanner.Scan() {
			fmt.Fprintln(os.Stderr, errScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting " + executable + " wrapper", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for " + executable + " wrapper to finish", err)
		os.Exit(1)
	}

	// all good can finish now
	os.Exit(0)

}
