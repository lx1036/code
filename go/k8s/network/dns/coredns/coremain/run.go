package coremain

import (
	"flag"
	"fmt"
	"github.com/caddyserver/caddy"
	"io/ioutil"
	"log"
	"os"
)

// Various CoreDNS constants.
const (
	CoreVersion = "1.6.4"
	coreName    = "CoreDNS"
	serverType  = "dns"
)

// Flags that control program flow or startup
var (
	conf    string
	version bool
	plugins bool
)

// Build information obtained with the help of -ldflags
var (
	appVersion = "(untracked dev build)" // inferred at startup
	devBuild   = true                    // inferred at startup

	buildDate        string // date -u
	gitTag           string // git describe --exact-match HEAD 2> /dev/null
	gitNearestTag    string // git describe --abbrev=0 --tags HEAD
	gitShortStat     string // git diff-index --shortstat
	gitFilesModified string // git diff-index --name-only HEAD

	// Gitcommit contains the commit where we built CoreDNS from.
	GitCommit string
)

// flagsBlacklist removes flags with these names from our flagset.
var flagsBlacklist = map[string]struct{}{
	"logtostderr":      {},
	"alsologtostderr":  {},
	"v":                {},
	"stderrthreshold":  {},
	"vmodule":          {},
	"log_backtrace_at": {},
	"log_dir":          {},
}

var flagsToKeep []*flag.Flag

func init() {
	caddy.DefaultConfigFile = "Corefile"
	caddy.Quiet = true // don't show init stuff from caddy

	flag.StringVar(&conf, "conf", "", "Corefile to load (default \""+caddy.DefaultConfigFile+"\")")
	flag.BoolVar(&plugins, "plugins", false, "List installed plugins")
	flag.StringVar(&caddy.PidFile, "pidfile", "", "Path to write pid file")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.BoolVar(&dnsserver.Quiet, "quiet", false, "Quiet mode (no initialization output)")

	caddy.RegisterCaddyfileLoader("flag", caddy.LoaderFunc(confLoader))
	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(defaultLoader))

	caddy.AppName = coreName
	caddy.AppVersion = CoreVersion
}

func Run() {
	caddy.TrapSignals()
	// Reset flag.CommandLine to get rid of unwanted flags for instance from glog (used in kubernetes).
	// And read the ones we want to keep.
	flag.VisitAll(func(f *flag.Flag) {
		if _, ok := flagsBlacklist[f.Name]; ok {
			return
		}
		flagsToKeep = append(flagsToKeep, f)
	})
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	for _, f := range flagsToKeep {
		flag.Var(f.Value, f.Name, f.Usage)
	}

	flag.Parse()

	if len(flag.Args()) > 0 {
		mustLogFatal(fmt.Errorf("extra command line arguments: %s", flag.Args()))
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(0) // Set to 0 because we're doing our own time, with timezone

	if plugins {
		fmt.Println(caddy.DescribePlugins())
		os.Exit(0)
	}
	// Get Corefile input
	corefile, err := caddy.LoadCaddyfile(serverType)
	if err != nil {
		mustLogFatal(err)
	}
	// Start your engines
	instance, err := caddy.Start(corefile)
	if err != nil {
		mustLogFatal(err)
	}

	instance.Wait()
}

// confLoader loads the Caddyfile using the -conf flag.
func confLoader(serverType string) (caddy.Input, error) {
	if conf == "" {
		return nil, nil
	}

	if conf == "stdin" {
		return caddy.CaddyfileFromPipe(os.Stdin, serverType)
	}

	contents, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       conf,
		ServerTypeName: serverType,
	}, nil
}

// defaultLoader loads the Corefile from the current working directory.
func defaultLoader(serverType string) (caddy.Input, error) {
	contents, err := ioutil.ReadFile(caddy.DefaultConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil
}

func mustLogFatal(args ...interface{}) {
	if !caddy.IsUpgrade() {
		log.SetOutput(os.Stderr)
	}
	log.Fatal(args...)
}