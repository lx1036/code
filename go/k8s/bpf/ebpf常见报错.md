# 1. "load program: no space left on device"

这种原因一般都是 ebpf verifier log size 过小，默认是 DefaultVerifierLogSize = 64 * 1024 = 65536，而 bpf c 程序里打印日志
bpf_printk() 比较多。解决办法可以在 load program/map 时设置 LogSize:

```go
    // Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	opts := &ebpf.CollectionOptions{
		Programs: ebpf.ProgramOptions{
			LogLevel: ebpf.LogLevelInstruction,
			LogSize:  64 * 1024 * 1024,
		},
	}
	if err := loadBpfObjects(&objs, opts); err != nil {
		logrus.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()
```


