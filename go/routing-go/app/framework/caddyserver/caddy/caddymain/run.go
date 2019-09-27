package caddymain

import (
    "bufio"
    "flag"
    "fmt"
    "github.com/klauspost/cpuid"
    "github.com/mholt/certmagic"
    lumberjack "gopkg.in/natefinch/lumberjack.v2"
    "io"
    "io/ioutil"
    "k8s-lx1036/routing-go/app/framework/caddyserver"
    "k8s-lx1036/routing-go/app/framework/caddyserver/telemetry"
    "log"
    "os"
    "path/filepath"
    "runtime"
    "runtime/debug"
    "strings"
)

/**
https://github.com/caddyserver/caddy/blob/master/caddy/caddymain/run.go
 */

var (
    disabledMetrics string
    serverType      string
    conf            string
    cpu             string
    envFile         string
    fromJSON        bool
    logfile         string
    logRollMB       int
    logRollCompress bool
    revoke          string
    toJSON          bool
    version         bool
    plugins         bool
    printEnv        bool
    validate        bool
)

// EnableTelemetry defines whether telemetry is enabled in Run.
var EnableTelemetry = true

const appName = "Caddy"


func init()  {
    caddy.TrapSignals()

    flag.BoolVar(&certmagic.Default.Agreed, "agree", false, "Agree to the CA's Subscriber Agreement")
    flag.StringVar(&certmagic.Default.CA, "ca", certmagic.Default.CA, "URL to certificate authority's ACME server directory")
    flag.StringVar(&certmagic.Default.DefaultServerName, "default-sni", certmagic.Default.DefaultServerName, "If a ClientHello ServerName is empty, use this ServerName to choose a TLS certificate")
    flag.BoolVar(&certmagic.Default.DisableHTTPChallenge, "disable-http-challenge", certmagic.Default.DisableHTTPChallenge, "Disable the ACME HTTP challenge")
    flag.BoolVar(&certmagic.Default.DisableTLSALPNChallenge, "disable-tls-alpn-challenge", certmagic.Default.DisableTLSALPNChallenge, "Disable the ACME TLS-ALPN challenge")
    flag.StringVar(&disabledMetrics, "disabled-metrics", "", "Comma-separated list of telemetry metrics to disable")
    flag.StringVar(&conf, "conf", "", "Caddyfile to load (default \""+caddy.DefaultConfigFile+"\")")
    flag.StringVar(&cpu, "cpu", "100%", "CPU cap")
    flag.BoolVar(&printEnv, "env", false, "Enable to print environment variables")
    flag.StringVar(&envFile, "envfile", "", "Path to file with environment variables to load in KEY=VALUE format")
    flag.BoolVar(&fromJSON, "json-to-caddyfile", false, "From JSON stdin to Caddyfile stdout")
    flag.BoolVar(&plugins, "plugins", false, "List installed plugins")
    flag.StringVar(&certmagic.Default.Email, "email", "", "Default ACME CA account email address")
    flag.DurationVar(&certmagic.HTTPTimeout, "catimeout", certmagic.HTTPTimeout, "Default ACME CA HTTP timeout")
    flag.StringVar(&logfile, "log", "", "Process log file")
    flag.IntVar(&logRollMB, "log-roll-mb", 100, "Roll process log when it reaches this many megabytes (0 to disable rolling)")
    flag.BoolVar(&logRollCompress, "log-roll-compress", true, "Gzip-compress rolled process log files")
    flag.StringVar(&caddy.PidFile, "pidfile", "", "Path to write pid file")
    flag.BoolVar(&caddy.Quiet, "quiet", false, "Quiet mode (no initialization output)")
    flag.StringVar(&revoke, "revoke", "", "Hostname for which to revoke the certificate")
    flag.StringVar(&serverType, "type", "http", "Type of server to run")
    flag.BoolVar(&toJSON, "caddyfile-to-json", false, "From Caddyfile stdin to JSON stdout")
    flag.BoolVar(&version, "version", false, "Show version")
    flag.BoolVar(&validate, "validate", false, "Parse the Caddyfile but do not start the server")
}

func Run()  {
    flag.Parse()
    module := getBuildModule()
    cleanModVersion := strings.TrimPrefix(module.Version, "v")

    caddy.AppName = appName
    caddy.AppVersion = module.Version
    certmagic.UserAgent = appName + "/" + cleanModVersion

    switch logfile {
    case "stdout":
        log.SetOutput(os.Stdout)
    case "stderr":
        log.SetOutput(os.Stderr)
    case "":
        log.SetOutput(ioutil.Discard)
    default:
        if logRollMB > 0 {
            log.SetOutput(&lumberjack.Logger{
                Filename:   logfile,
                MaxSize:    logRollMB,
                MaxAge:     14,
                MaxBackups: 10,
                Compress:   logRollCompress,
            })
        } else {
            err := os.MkdirAll(filepath.Dir(logfile), 0755)
            if err != nil {
                mustLogFatalf("%v", err)
            }
            f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
            if err != nil {
                mustLogFatalf("%v", err)
            }
            // don't close file; log should be writeable for duration of process
            log.SetOutput(f)
        }
    }

    // load all additional envs as soon as possible
    if err := LoadEnvFromFile(envFile); err != nil {
        mustLogFatalf("%v", err)
    }

    // initialize telemetry client
    if EnableTelemetry {
        err := initTelemetry()
        if err != nil {
            mustLogFatalf("[ERROR] Initializing telemetry: %v", err)
        }
    } else if disabledMetrics != "" {
        mustLogFatalf("[ERROR] Cannot disable specific metrics because telemetry is disabled")
    }

    // Check for one-time actions
    if revoke != "" {
        err := caddytls.Revoke(revoke)
        if err != nil {
            mustLogFatalf("%v", err)
        }
        fmt.Printf("Revoked certificate for %s\n", revoke)
        os.Exit(0)
    }

    if version {
        if module.Sum != "" {
            // a build with a known version will also have a checksum
            fmt.Printf("Caddy %s (%s)\n", module.Version, module.Sum)
        } else {
            fmt.Println(module.Version)
        }
        os.Exit(0)
    }

    if plugins {
        fmt.Println(caddy.DescribePlugins())
        os.Exit(0)
    }

    // Check if we just need to do a Caddyfile Convert and exit
    checkJSONCaddyfile()

    // Set CPU cap
    err := setCPU(cpu)
    if err != nil {
        mustLogFatalf("%v", err)
    }

    // Executes Startup events
    caddy.EmitEvent(caddy.StartupEvent, nil)

    // Get Caddyfile input
    caddyfileinput, err := caddy.LoadCaddyfile(serverType)
    if err != nil {
        mustLogFatalf("%v", err)
    }

    if validate {
        err := caddy.ValidateAndExecuteDirectives(caddyfileinput, nil, true)
        if err != nil {
            mustLogFatalf("%v", err)
        }
        msg := "Caddyfile is valid"
        fmt.Println(msg)
        log.Printf("[INFO] %s", msg)
        os.Exit(0)
    }

    // Log Caddy version before start
    log.Printf("[INFO] Caddy version: %s", module.Version)

    // Start your engines
    instance, err := caddy.Start(caddyfileinput)
    if err != nil {
        mustLogFatalf("%v", err)
    }

    // Begin telemetry (these are no-ops if telemetry disabled)
    telemetry.Set("caddy_version", module.Version)
    telemetry.Set("num_listeners", len(instance.Servers()))
    telemetry.Set("server_type", serverType)
    telemetry.Set("os", runtime.GOOS)
    telemetry.Set("arch", runtime.GOARCH)
    telemetry.Set("cpu", struct {
        BrandName  string `json:"brand_name,omitempty"`
        NumLogical int    `json:"num_logical,omitempty"`
        AESNI      bool   `json:"aes_ni,omitempty"`
    }{
        BrandName:  cpuid.CPU.BrandName,
        NumLogical: runtime.NumCPU(),
        AESNI:      cpuid.CPU.AesNi(),
    })
    if containerized := detectContainer(); containerized {
        telemetry.Set("container", containerized)
    }
    telemetry.StartEmitting()

    // Twiddle your thumbs
    instance.Wait()
}

// mustLogFatalf wraps log.Fatalf() in a way that ensures the
// output is always printed to stderr so the user can see it
// if the user is still there, even if the process log was not
// enabled. If this process is an upgrade, however, and the user
// might not be there anymore, this just logs to the process
// log and exits.
func mustLogFatalf(format string, args ...interface{}) {
    if !caddy.IsUpgrade() {
        log.SetOutput(os.Stderr)
    }
    log.Fatalf(format, args...)
}

// LoadEnvFromFile loads additional envs if file provided and exists
// Envs in file should be in KEY=VALUE format
func LoadEnvFromFile(envFile string) error {
    if envFile == "" {
        return nil
    }

    file, err := os.Open(envFile)
    if err != nil {
        return err
    }
    defer file.Close()

    envMap, err := ParseEnvFile(file)
    if err != nil {
        return err
    }

    for k, v := range envMap {
        if err := os.Setenv(k, v); err != nil {
            return err
        }
    }

    return nil
}

// ParseEnvFile implements parse logic for environment files
func ParseEnvFile(envInput io.Reader) (map[string]string, error) {
    envMap := make(map[string]string)

    scanner := bufio.NewScanner(envInput)
    var line string
    lineNumber := 0

    for scanner.Scan() {
        line = strings.TrimSpace(scanner.Text())
        lineNumber++

        // skip lines starting with comment
        if strings.HasPrefix(line, "#") {
            continue
        }

        // skip empty line
        if len(line) == 0 {
            continue
        }

        fields := strings.SplitN(line, "=", 2)
        if len(fields) != 2 {
            return nil, fmt.Errorf("can't parse line %d; line should be in KEY=VALUE format", lineNumber)
        }

        if strings.Contains(fields[0], " ") {
            return nil, fmt.Errorf("can't parse line %d; KEY contains whitespace", lineNumber)
        }

        key := fields[0]
        val := fields[1]

        if key == "" {
            return nil, fmt.Errorf("can't parse line %d; KEY can't be empty string", lineNumber)
        }
        envMap[key] = val
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return envMap, nil
}

// getBuildModule returns the build info of Caddy
// from debug.BuildInfo (requires Go modules). If
// no version information is available, a non-nil
// value will still be returned, but with an
// unknown version.
func getBuildModule() *debug.Module {
    bi, ok := debug.ReadBuildInfo()
    if ok {
        // The recommended way to build Caddy involves
        // creating a separate main module, which
        // preserves caddy a read-only dependency
        // TODO: track related Go issue: https://github.com/golang/go/issues/29228
        for _, mod := range bi.Deps {
            if mod.Path == "github.com/caddyserver/caddy" {
                return mod
            }
        }
    }
    return &debug.Module{Version: "unknown"}
}

// Check if we just need to do a Caddyfile Convert and exit
func checkJSONCaddyfile() {
    if fromJSON {
        jsonBytes, err := ioutil.ReadAll(os.Stdin)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Read stdin failed: %v", err)
            os.Exit(1)
        }
        caddyfileBytes, err := caddyfile.FromJSON(jsonBytes)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Converting from JSON failed: %v", err)
            os.Exit(2)
        }
        fmt.Println(string(caddyfileBytes))
        os.Exit(0)
    }
    if toJSON {
        caddyfileBytes, err := ioutil.ReadAll(os.Stdin)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Read stdin failed: %v", err)
            os.Exit(1)
        }
        jsonBytes, err := caddyfile.ToJSON(caddyfileBytes)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Converting to JSON failed: %v", err)
            os.Exit(2)
        }
        fmt.Println(string(jsonBytes))
        os.Exit(0)
    }
}
