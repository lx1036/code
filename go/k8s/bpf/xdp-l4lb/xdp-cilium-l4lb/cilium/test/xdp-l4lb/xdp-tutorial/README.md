

编写 bpf 程序步骤如下：
* 1.编写 bpf.c 文件和 bpf.go 文件
* 2.go generate 生成 bpf.o 文件 和 bpfeb.go 文件
* 3.main.go 里写业务逻辑(bpf.c 不要和 bpf.go 文件在同一个文件夹)

