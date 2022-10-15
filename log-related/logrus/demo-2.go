package main

import (
	"fmt"
	"main/logrus"
	"os"
)

// Create a new instance of the logger. You can have any number of instances.
var log = logrus.New()

type hook struct {
	name string
}

func (h hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.InfoLevel,
	}
}

func (h hook) Fire(entry *logrus.Entry) error {
	fmt.Printf("in [%s] hook, entry[%v]\n", h.name, entry)
	return nil
}

func main() {
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	log.Out = os.Stdout

	// You could set this to any `io.Writer` such as a file
	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err == nil {
	//  log.Out = file
	// } else {
	//  log.Info("Failed to log to file, using default stderr")
	// }

	log.AddHook(hook{
		name: "hook-name",
	})

	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")
}
