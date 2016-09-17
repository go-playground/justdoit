package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"sync"

	"path/filepath"
	"regexp"

	"gopkg.in/fsnotify.v1"

	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
)

func init() {
	cLog := console.New()
	cLog.RedirectSTDLogOutput(true)
	log.RegisterHandler(cLog, log.AllLevels...)
}

var (
	lock    sync.Mutex
	running bool
	proc    *exec.Cmd
)

func main() {

	flagWatch := flag.String("watch", "./", "Directory to watch for changes (recursive)")
	flagExclude := flag.String("exclude", "(.git|vendor)$", "Regex of paths to exclude")
	flagInclude := flag.String("include", `(.+\.go|.+\.c)$`, "Regex of files to include")
	flagBuild := flag.String("build", "go install -v", "Command to Build/Compile program")
	flagRun := flag.String("run", "", "Command to run your application")

	flag.Parse()

	if len(strings.TrimSpace(*flagBuild)) == 0 {
		log.Fatal("build is a required argument")
	}

	if len(strings.TrimSpace(*flagRun)) == 0 {
		log.Fatal("run is a required argument")
	}

	var include *regexp.Regexp
	var exclude *regexp.Regexp

	absWatch, err := filepath.Abs(*flagWatch)
	if err != nil {
		log.WithFields(log.F("error", err)).Fatal("invalid watch directory, could not determine absolute path")
	}

	fi, err := os.Stat(absWatch)
	if err != nil || !fi.IsDir() {
		log.WithFields(log.F("error", err)).Fatal("invalid watch directory")
	}

	if len(*flagInclude) > 0 {
		include, err = regexp.Compile(*flagInclude)
		if err != nil {
			log.WithFields(log.F("error", err)).Fatal("invalid include regex")
		}
	}

	if len(*flagExclude) > 0 {
		exclude, err = regexp.Compile(*flagExclude)
		if err != nil {
			log.WithFields(log.F("error", err)).Fatal("invalid include regex")
		}
	}

	notif := make(chan struct{})

	go watch(notif, absWatch, include, exclude)

	// trigger event for initial run of application
	go func() {
		notif <- struct{}{}
	}()

	build(*flagBuild, *flagRun, notif)
}

func build(buildCmd, executeCms string, event <-chan struct{}) {

	buildArgs := strings.Split(buildCmd, " ")
	executeArgs := strings.Split(executeCms, " ")

	for range event {
		log.Notice("Running build command")
		if !execute(buildArgs, false) {
			kill(true)
			continue
		}

		go run(executeArgs)
	}
}

func kill(unlock bool) {
	lock.Lock()

	if running {

		// kill process

		err := proc.Process.Kill()
		if err != nil {
			log.WithFields(log.F("error", err)).Error("could not kill process")
		}

		running = false
	}

	if unlock {
		lock.Unlock()
	}
}

func run(args []string) {

	kill(false)
	execute(args, true)
}

// runs generically provided command
func execute(args []string, setProc bool) (success bool) {

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if setProc {
		proc = cmd
		running = true
		lock.Unlock()
		log.Notice("Executing run command")
	}

	err := cmd.Run()
	if !setProc && err != nil { // not outputting killed for already running
		log.WithFields(log.F("error", err)).Notice("error stopping cmd")
		return false
	}

	return true
}

func watch(notif chan<- struct{}, watch string, include, exclude *regexp.Regexp) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithFields(log.F("error", err)).Fatal("issue creating watcher")
	}

	defer watcher.Close()

	walker := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {

			if exclude != nil && exclude.MatchString(path) {
				return filepath.SkipDir
			}

			err = watcher.Add(path)
			if err != nil {
				log.WithFields(log.F("error", err)).Warn("issue adding directory to watch")
			}
		}

		return nil
	}

	err = filepath.Walk(watch, walker)
	if err != nil {
		log.WithFields(log.F("error", err)).Fatal("could not walk watch path")
	}

	// evt := make(chan struct{})

	// go func() {
	// 	for {
	// 		<-evt

	// 	SELECT:
	// 		select {
	// 		case <-time.After(700 * time.Millisecond):
	// 			notif <- struct{}{}
	// 		case <-evt:
	// 			goto SELECT
	// 		}
	// 	}
	// }()

	go func() {
		for {
			select {
			case event := <-watcher.Events:

				if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {

					if include != nil && !include.MatchString(event.Name) {
						continue
					}

					notif <- struct{}{}
					// evt <- struct{}{}
				}

			case err := <-watcher.Errors:
				log.WithFields(log.F("error", err)).Error("watcher error")
			}
		}
	}()

	<-make(chan bool)
}
