package exec

import (
    "bufio"
    "bytes"
    "context"
    "fmt"
    "github.com/sirupsen/logrus"
    "os/exec"
    "strings"
)

// Cmd wraps exec.Cmd with a context to provide convenient execution of a
// command with nice checking of the context timeout in the form:
//
// err := exec.Prog().WithTimeout(5*time.Second, myprog, myargs...).CombinedOutput(log, verbose)
type Cmd struct {
    *exec.Cmd
    ctx      context.Context
    cancelFn func()

    // filters is a slice of strings that should be omitted from logging.
    filters []string
}

func (c *Cmd) CombinedOutput(scopedLog *logrus.Entry, verbose bool) ([]byte, error) {
    out, err := combinedOutput(c.ctx, c.Cmd, c.filters, scopedLog, verbose)
    if c.cancelFn != nil {
        c.cancelFn()
    }
    return out, err
}

func warnToLog(cmd *exec.Cmd, filters []string, out []byte, scopedLog *logrus.Entry, err error) {
    scopedLog.WithError(err).WithField("cmd", cmd.Args).Error("Command execution failed")
    scanner := bufio.NewScanner(bytes.NewReader(out))
scan:
    for scanner.Scan() {
        text := scanner.Text()
        for _, filter := range filters {
            if strings.Contains(text, filter) {
                continue scan
            }
        }
        scopedLog.Warn(text)
    }
}

// combinedOutput is the core implementation of catching deadline exceeded
// options and logging errors, with an optional set of filtered outputs.
func combinedOutput(ctx context.Context, cmd *exec.Cmd, filters []string, scopedLog *logrus.Entry, verbose bool) ([]byte, error) {
    out, err := cmd.CombinedOutput()
    if ctx.Err() != nil {
        scopedLog.WithError(err).WithField("cmd", cmd.Args).Error("Command execution failed")
        return nil, fmt.Errorf("command execution failed for %s: %s", cmd.Args, ctx.Err())
    }
    if err != nil && verbose {
        warnToLog(cmd, filters, out, scopedLog, err)
    }
    return out, err
}

// CommandContext wraps exec.CommandContext to allow this package to be used as
// a drop-in replacement for the standard exec library.
func CommandContext(ctx context.Context, prog string, args ...string) *Cmd {
    return &Cmd{
        Cmd: exec.CommandContext(ctx, prog, args...),
        ctx: ctx,
    }
}

func WithCancel(ctx context.Context, prog string, args ...string) (*Cmd, context.CancelFunc) {
    newCtx, cancel := context.WithCancel(ctx)
    cmd := CommandContext(newCtx, prog, args...)
    return cmd, cancel
}
