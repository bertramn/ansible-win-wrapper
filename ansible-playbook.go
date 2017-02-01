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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/vaughan0/go-ini"
)

func main() {

	var wrapperExecutableName = filepath.Base(os.Args[0])
	var extension = filepath.Ext(wrapperExecutableName)
	var executable = wrapperExecutableName[0 : len(wrapperExecutableName)-len(extension)]

	f, err := os.OpenFile(executable+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating "+executable+" log file", err)
		os.Exit(1)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("*******************************************************")

	log.Printf("Target executable: %s\n", executable)
	wrapperExecutableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting path from "+os.Args[0], err)
		os.Exit(1)
	}
	log.Printf("Wrapper dir: %s\n", wrapperExecutableDir)

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting current working folder", err)
		os.Exit(1)
	}
	log.Printf("Current dir: %s\n", pwd)

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
	log.Printf("Cygwin Home: %s\n", cygHome)

	// // environment variables
	// log.Println("Environment variables:")
	// for i, e := range os.Environ() {
	// 	pair := strings.Split(e, "=")
	// 	log.Printf(" > %d: %s=%s\n", i, pair[0], pair[1])
	// }

	// args provided
	log.Println("System args:")
	args := make([]string, len(os.Args[1:])+1, len(os.Args[1:])+1)
	args[0] = "/bin/" + executable
	for i, a := range os.Args[1:] {
		log.Printf(" > %d: [%-10s] %s\n", i, reflect.TypeOf(a), a)
		if strings.HasPrefix(a, "--extra-vars={") {
			var sarg = strings.Split(a, "=")
			args[i+1] = fmt.Sprintf("%s=%s", sarg[0], sarg[1])
			log.Printf("extra vars set to: %s", args[i+1])
		} else {
			args[i+1] = a
		}
	}

	var bashCmd = filepath.ToSlash(cygHome + "/bin/python2.7.exe")
	log.Printf("execute\n  %s %s", bashCmd, strings.Join(args[:], " "))
	cmd := exec.Command(bashCmd, args...)
	// echo stdin back to console
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating "+executable+" wrapper stdout pipeline", err)
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
		fmt.Fprintln(os.Stderr, "Error creating "+executable+" wrapper stderr pipeline", err)
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
