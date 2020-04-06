package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/containerd/fifo"
	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/jsonfilelog"
	"github.com/docker/go-plugins-helpers/sdk"
	protoio "github.com/gogo/protobuf/io"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type StartLoggingRequest struct {
	File string
	Info logger.Info
}
type StopLoggingRequest struct {
	File string
}
type response struct {
	Err string
}

func respond(err error, w http.ResponseWriter) {
	var res response
	if err != nil {
		res.Err = err.Error()
	}
	_ = json.NewEncoder(w).Encode(&res)
}

/**
@see https://github.com/cpuguy83/docker-log-driver-test
*/
func main() {
	logrus.SetLevel(logrus.DebugLevel)

	handler := sdk.NewHandler(`{Implements:["LoggingDriver"]}`)

	driver := NewDriver()

	handler.HandleFunc("/LogDriver.StartLogging", func(w http.ResponseWriter, r *http.Request) {
		// https://docs.docker.com/engine/extend/plugins_logging/#logdriverstartlogging
		var req StartLoggingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Info.ContainerID == "" {
			respond(errors.New("must provide container id in log context"), w)
			return
		}

		err := driver.StartLogging(req.File, req.Info)
		respond(err, w)
	})

	handler.HandleFunc("/LogDriver.StopLogging", func(w http.ResponseWriter, r *http.Request) {
		// https://docs.docker.com/engine/extend/plugins_logging/#logdriverstoplogging
		var req StopLoggingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := driver.StopLogging(req.File)
		respond(err, w)
	})

	handler.HandleFunc("/LogDriver.Capabilities", func(w http.ResponseWriter, r *http.Request) {

	})

	handler.HandleFunc("/LogDriver.ReadLogs", func(w http.ResponseWriter, r *http.Request) {

	})

	if err := handler.ServeUnix("jsonfile", 0); err != nil {
		panic(err)
	}
}

type Driver struct {
	mu     sync.Mutex
	logs   map[string]*logPair
	idx    map[string]*logPair
	logger logger.Logger
}
type logPair struct {
	l      logger.Logger
	stream io.ReadCloser
	info   logger.Info
}

func NewDriver() *Driver {
	return &Driver{
		logs: map[string]*logPair{},
		idx:  map[string]*logPair{},
	}
}

//{
//	"File": "/path/to/file/stream",
//	"Info": {
//		"ContainerID": "123456"
//	}
//}
func (driver *Driver) StartLogging(file string, info logger.Info) error {
	driver.mu.Lock()
	if _, exists := driver.logs[file]; exists {
		driver.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	driver.mu.Unlock()

	if info.LogPath == "" {
		info.LogPath = filepath.Join("/var/log/docker", info.ContainerID)
	}
	if err := os.MkdirAll(filepath.Dir(info.LogPath), 0755); err != nil {
		return errors.New(fmt.Sprintf("error setting up logger dir: %s", err.Error()))
	}

	jsonfile, err := jsonfilelog.New(info)
	if err != nil {
		return errors.New(fmt.Sprintf("error creating jsonfile logger: %s", err.Error()))
	}

	logrus.WithField("id", info.ContainerID).
		WithField("file", file).
		WithField("logpath", info.LogPath).
		Debug("Start logging")
	f, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.New(fmt.Sprintf("error opening logger fifo: %s", err.Error()))
	}

	driver.mu.Lock()
	pair := &logPair{jsonfile, f, info}
	driver.logs[file] = pair
	driver.idx[info.ContainerID] = pair
	driver.mu.Unlock()

	go consumeLog(pair)

	return nil
}

func consumeLog(pair *logPair) {
	dec := protoio.NewUint32DelimitedReader(pair.stream, binary.BigEndian, 1e6)
	defer dec.Close()

	var buf logdriver.LogEntry
	for {
		if err := dec.ReadMsg(&buf); err != nil {
			if err == io.EOF {

			}

			continue
		}

		var msg logger.Message
		msg.Line = buf.Line
		msg.Source = buf.Source
		if buf.PartialLogMetadata != nil {
			msg.PLogMetaData.ID = buf.PartialLogMetadata.Id
			msg.PLogMetaData.Last = buf.PartialLogMetadata.Last
			msg.PLogMetaData.Ordinal = int(buf.PartialLogMetadata.Ordinal)
		}
		msg.Timestamp = time.Unix(0, buf.TimeNano)
		if err := pair.l.Log(&msg); err != nil {
			logrus.WithField("id", pair.info.ContainerID).
				WithField("message", msg).
				Errorf("error writing log message: %s", err.Error())
			continue
		}

		buf.Reset()
	}
}

func (driver *Driver) StopLogging(file string) error {
	logrus.WithField("file", file).Debug("Stop logging")

	driver.mu.Lock()
	pair, ok := driver.logs[file]
	if ok {
		pair.stream.Close()
		delete(driver.logs, file)
	}
	driver.mu.Unlock()

	return nil
}
