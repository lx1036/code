package cmd

import (
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/link"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/defaults"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent/listener"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/format"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/payload"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	connTimeout = 12 * time.Second
)

// monitorCmd represents the monitor command
var (
	monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Display BPF program events",
		Long: `The monitor displays notifications and events emitted by the BPF
programs attached to endpoints and devices. This includes:
  * Dropped packet notifications
  * Captured packet traces
  * Policy verdict notifications
  * Debugging information`,
		Run: func(cmd *cobra.Command, args []string) {
			runMonitor(args)
		},
	}
	linkCache  = link.NewLinkCache()
	printer    = format.NewMonitorFormatter(format.INFO, linkCache)
	socketPath = ""
	verbosity  = []bool{}
)

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().BoolVar(&printer.Hex, "hex", false, "Do not dissect, print payload in HEX")
	monitorCmd.Flags().VarP(&printer.EventTypes, "type", "t", fmt.Sprintf("Filter by event types %v", monitor.GetAllTypes()))
	monitorCmd.Flags().Var(&printer.FromSource, "from", "Filter by source endpoint id")
	monitorCmd.Flags().Var(&printer.ToDst, "to", "Filter by destination endpoint id")
	monitorCmd.Flags().Var(&printer.Related, "related-to", "Filter by either source or destination endpoint id")
	monitorCmd.Flags().BoolSliceVarP(&verbosity, "verbose", "v", nil, "Enable verbose output (-v, -vv)")
	monitorCmd.Flags().Lookup("verbose").NoOptDefVal = "false"
	monitorCmd.Flags().BoolVarP(&printer.JSONOutput, "json", "j", false, "Enable json output. Shadows -v flag")
	monitorCmd.Flags().BoolVarP(&printer.Numeric, "numeric", "n", false, "Display all security identities as numeric values")
	monitorCmd.Flags().StringVar(&socketPath, "monitor-socket", "", "Configure monitor socket path")
	viper.BindEnv("monitor-socket", "CILIUM_MONITOR_SOCK")
	viper.BindPFlags(monitorCmd.Flags())
}

func runMonitor(args []string) {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "Error: arguments not recognized")
		os.Exit(1)
	}

	//validateEndpointsFilters()
	//setVerbosity()
	setupSigHandler()

	// On EOF, retry
	// On other errors, exit
	// always wait connTimeout when retrying
	for ; ; time.Sleep(connTimeout) {
		conn, version, err := openMonitorSock(viper.GetString("monitor-socket"))
		if err != nil {
			log.WithError(err).Error("Cannot open monitor socket")
			return
		}

		err = consumeMonitorEvents(conn, version)
		switch {
		case err == nil:
		// no-op

		case err == io.EOF, errors.Is(err, io.ErrUnexpectedEOF):
			log.WithError(err).Warn("connection closed")
			continue

		default:
			log.WithError(err).Fatal("decoding error")
		}
	}
}

// openMonitorSock attempts to open a version specific monitor socket It
// returns a connection, with a version, or an error.
func openMonitorSock(path string) (conn net.Conn, version listener.Version, err error) {
	errs := make([]string, 0)

	// try the user-provided socket
	if path != "" {
		conn, err = net.Dial("unix", path)
		if err == nil {
			version = listener.Version1_2
			return conn, version, nil
		}
		errs = append(errs, path+": "+err.Error())
	}

	// try the 1.2 socket
	conn, err = net.Dial("unix", defaults.MonitorSockPath1_2)
	if err == nil {
		return conn, listener.Version1_2, nil
	}
	errs = append(errs, defaults.MonitorSockPath1_2+": "+err.Error())

	return nil, listener.VersionUnsupported, fmt.Errorf("cannot find or open a supported node-monitor socket. %s", strings.Join(errs, ","))
}

func setupSigHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Fprintf(os.Stderr, "\nReceived an interrupt, disconnecting from monitor...\n\n")
			os.Exit(0) // INFO: 主进程会退出
		}
	}()
}

// consumeMonitorEvents handles and prints events on a monitor connection. It
// calls getMonitorParsed to construct a monitor-version appropriate parser.
// It closes conn on return, and returns on error, including io.EOF
func consumeMonitorEvents(conn net.Conn, version listener.Version) error {
	defer conn.Close()

	getParsedPayload, err := getMonitorParser(conn, version)
	if err != nil {
		return err
	}

	for {
		pl, err := getParsedPayload()
		if err != nil {
			return err
		}
		if !printer.FormatEvent(pl) {
			// earlier code used an else to handle this case, along with pl.Type ==
			// payload.RecordLost above. It should be safe to call lostEvent to match
			// the earlier behaviour, despite it not being wholly correct.
			log.WithError(err).WithField("type", pl.Type).Warn("Unknown payload type")
			format.LostEvent(pl.Lost, pl.CPU)
		}
	}
}

type eventParserFunc func() (*payload.Payload, error)

// getMonitorParser constructs and returns an eventParserFunc. It is
// appropriate for the monitor API version passed in.
func getMonitorParser(conn net.Conn, version listener.Version) (parser eventParserFunc, err error) {
	switch version {
	case listener.Version1_2:
		var (
			pl  payload.Payload
			dec = gob.NewDecoder(conn)
		)
		// This implemenents the newer 1.2 API. Each listener maintains its own gob
		// session, and type information is only ever sent once.
		return func() (*payload.Payload, error) {
			if err := pl.DecodeBinary(dec); err != nil {
				return nil, err
			}
			return &pl, nil
		}, nil

	default:
		return nil, fmt.Errorf("unsupported version %s", version)
	}
}
