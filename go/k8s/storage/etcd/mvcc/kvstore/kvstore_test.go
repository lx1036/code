package kvstore

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestKVStore(test *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	commitC := make(chan *string)
	errorC := make(chan error)
	readCommits(commitC, errorC)

	str := "aaa"
	err := errors.New("shutdown")

	go func() {
		commitC <- &str
		errorC <- err
	}()
}

func readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			log.Info("ok empty")
		}

		log.Info("ok")
	}

	if err, ok := <-errorC; ok {
		log.Error(err)
	}
}
