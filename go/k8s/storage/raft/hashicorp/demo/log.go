package main

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"k8s.io/klog/v2"
)

type klogger struct{}

func (k *klogger) Trace(msg string, args ...interface{}) {
	panic("implement me")
}

func (k *klogger) Debug(msg string, args ...interface{}) {
	klog.Infof(msg, args)
}

func (k *klogger) Info(msg string, args ...interface{}) {
	klog.Infof(msg, args)
}

func (k *klogger) Warn(msg string, args ...interface{}) {
	klog.Warningf(msg, args)
}

func (k *klogger) Error(msg string, args ...interface{}) {
	klog.Errorf(msg, args)
}

func (k *klogger) IsTrace() bool {
	panic("implement me")
}

func (k *klogger) IsDebug() bool {
	return true
}

func (k *klogger) IsInfo() bool {
	return true
}

func (k *klogger) IsWarn() bool {
	return true
}

func (k *klogger) IsError() bool {
	return true
}

func (k *klogger) With(args ...interface{}) hclog.Logger {
	panic("implement me")
}

func (k *klogger) Named(name string) hclog.Logger {
	panic("implement me")
}

func (k *klogger) ResetNamed(name string) hclog.Logger {
	panic("implement me")
}

func (k *klogger) SetLevel(level hclog.Level) {
	panic("implement me")
}

func (k *klogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	panic("implement me")
}

func (k *klogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	panic("implement me")
}

func NewLogger() *klogger {
	return &klogger{}
}
