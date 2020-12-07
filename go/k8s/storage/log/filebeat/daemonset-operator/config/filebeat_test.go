package config

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

type LogCollectorType string

const (
	Daemonset LogCollectorType = "daemonset"
	Sidecar   LogCollectorType = "sidecar"
)

type LogType string

const (
	Stdout LogType = "stdout"
	File   LogType = "file"
)

type FilebeatInput struct {
	LogCollectorType LogCollectorType `json:"log_collector_type"` //0: sidecar, 1: daemonset
	LogType          LogType          `json:"log_type,omitempty"` //0: stdout， 1: filelog
	Topic            string           `json:"topic"`
	Hosts            string           `json:"hosts"`
	Containers       []string         `json:"containers"`
	CustomField      string           `json:"custom_field"`

	Paths           []string `json:"paths,omitempty"` //only daemonset mode, and stdout paths will be nil
	MultilineEnable bool     `json:"multiline_enable"`
}

func TestFilebeatTemplate(test *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	tpl, err := template.ParseFiles("../pkg/controller/inputs.yml.template")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	type Data struct {
		FilebeatInputs []FilebeatInput `json:"FilebeatInputs"`
	}

	data := Data{
		FilebeatInputs: []FilebeatInput{
			{
				Hosts: "http://1.2.3.4",
				Paths: []string{
					"/var/lib/docker/containers/1/1-json.log",
					"/var/lib/docker/containers/2/2-json.log",
					"/var/lib/docker/containers/3/3-json.log",
				},
				Topic:       "topic_1",
				CustomField: "IDC=beijing",
			},
			{
				Hosts: "http://2.3.4.5",
				Paths: []string{
					"/var/lib/docker/containers/1/1-json.log",
					"/var/lib/docker/containers/2/2-json.log",
					"/var/lib/docker/containers/3/3-json.log",
				},
				Topic:       "topic_2",
				CustomField: "IDC=shanghai",
			},
		},
	}

	buf := bytes.NewBufferString("")
	if err = tpl.Execute(buf, data); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if err = ioutil.WriteFile("inputs.yml", buf.Bytes(), os.ModePerm); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
