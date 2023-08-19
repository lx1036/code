// Code generated by bpf2go; DO NOT EDIT.
//go:build 386 || amd64 || amd64p32 || arm || arm64 || loong64 || mips64le || mips64p32le || mipsle || ppc64le || riscv64

package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

type tcpHeaderHdrStg struct {
	Active    bool
	ResendSyn bool
	Syncookie bool
	Fastopen  bool
}

type tcpHeaderLinumErr struct {
	Linum uint32
	Err   int32
}

// loadTcpHeader returns the embedded CollectionSpec for tcpHeader.
func loadTcpHeader() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_TcpHeaderBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load tcpHeader: %w", err)
	}

	return spec, err
}

// loadTcpHeaderObjects loads tcpHeader and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*tcpHeaderObjects
//	*tcpHeaderPrograms
//	*tcpHeaderMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadTcpHeaderObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadTcpHeader()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// tcpHeaderSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type tcpHeaderSpecs struct {
	tcpHeaderProgramSpecs
	tcpHeaderMapSpecs
}

// tcpHeaderSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type tcpHeaderProgramSpecs struct {
	Estab *ebpf.ProgramSpec `ebpf:"estab"`
}

// tcpHeaderMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type tcpHeaderMapSpecs struct {
	HdrStgMap     *ebpf.MapSpec `ebpf:"hdr_stg_map"`
	LportLinumMap *ebpf.MapSpec `ebpf:"lport_linum_map"`
}

// tcpHeaderObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadTcpHeaderObjects or ebpf.CollectionSpec.LoadAndAssign.
type tcpHeaderObjects struct {
	tcpHeaderPrograms
	tcpHeaderMaps
}

func (o *tcpHeaderObjects) Close() error {
	return _TcpHeaderClose(
		&o.tcpHeaderPrograms,
		&o.tcpHeaderMaps,
	)
}

// tcpHeaderMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadTcpHeaderObjects or ebpf.CollectionSpec.LoadAndAssign.
type tcpHeaderMaps struct {
	HdrStgMap     *ebpf.Map `ebpf:"hdr_stg_map"`
	LportLinumMap *ebpf.Map `ebpf:"lport_linum_map"`
}

func (m *tcpHeaderMaps) Close() error {
	return _TcpHeaderClose(
		m.HdrStgMap,
		m.LportLinumMap,
	)
}

// tcpHeaderPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadTcpHeaderObjects or ebpf.CollectionSpec.LoadAndAssign.
type tcpHeaderPrograms struct {
	Estab *ebpf.Program `ebpf:"estab"`
}

func (p *tcpHeaderPrograms) Close() error {
	return _TcpHeaderClose(
		p.Estab,
	)
}

func _TcpHeaderClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed tcpheader_bpfel.o
var _TcpHeaderBytes []byte
