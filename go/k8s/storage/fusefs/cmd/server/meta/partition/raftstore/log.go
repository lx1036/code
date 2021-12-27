package raftstore

import (
	"github.com/tiglabs/raft/logger"
	"k8s.io/klog/v2"
)

func init() {
	logger.SetLogger(klogger{})
}

type klogger struct{}

func (k klogger) IsEnableDebug() bool {
	return true
}

func (k klogger) IsEnableInfo() bool {
	return true
}

func (k klogger) IsEnableWarn() bool {
	return true
}

func (k klogger) Debug(format string, v ...interface{}) {
	klog.Infof(format, v)
}

func (k klogger) Info(format string, v ...interface{}) {
	klog.Infof(format, v)
}

func (k klogger) Warn(format string, v ...interface{}) {
	klog.Warningf(format, v)
}

func (k klogger) Error(format string, v ...interface{}) {
	klog.Errorf(format, v)
}
