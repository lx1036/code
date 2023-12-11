package tcprtt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"syscall"
	"testing"
)

// go test -v -run ^TestTCPRtt$ .
func TestTCPRtt(test *testing.T) {
	raddr, err := net.ResolveTCPAddr("tcp", "www.baidu.com:80")
	if err != nil {
		logrus.Fatal(err)
	}
	// 153.3.238.110
	logrus.Info(raddr)

	ip, err := net.LookupIP("www.baidu.com")
	if err != nil {
		logrus.Fatal(err)
	}
	// [153.3.238.110 153.3.238.102 240e:e9:6002:15a:0:ff:b05c:1278 240e:e9:6002:15c:0:ff:b015:146f]
	logrus.Info(ip)
	sa := &unix.SockaddrInet4{
		Port: 80,
		Addr: [4]byte{},
	}
	copy(sa.Addr[:], ip[0])

	socket, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		test.Fatal(err)
	}
	getTCPInfo(socket)

	unix.Connect(socket, sa)
	getTCPInfo(socket)

	unix.Send(socket, []byte("hello"), 0)
	getTCPInfo(socket)
}

func getTCPInfo(socketFd int) {
	tcpInfo, err := unix.GetsockoptTCPInfo(socketFd, syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("%+v\n\n", tcpInfo)
}

func TestTCPInfo(test *testing.T) {
	//r, err := http.Get("baidu.com:80")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer r.Body.Close()
	//data, err := io.ReadAll(r.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//logrus.Info(data)
	//return

	// 建立一个 TCP 连接
	conn, err := net.Dial("tcp", "baidu.com:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 取得 TCP 连接的文件描述符
	tcpConn := conn.(*net.TCPConn)
	file, err := tcpConn.File()
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fd := int(file.Fd())

	// 获取 TCP 信息
	tcpInfo, err := unix.GetsockoptTCPInfo(fd, syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		log.Fatal(err)
	}
	/* &{State:1 Ca_state:0 Retransmits:0 Probes:0 Backoff:0 Options:6 Rto:252000 Ato:0 Snd_mss:1452 Rcv_mss:536
	Unacked:0 Sacked:0 Lost:0 Retrans:0 Fackets:0 Last_data_sent:0 Last_ack_sent:0 Last_data_recv:0 Last_ack_recv:0
	Pmtu:1500 Rcv_ssthresh:64076 Rtt:50632 Rttvar:25316 Snd_ssthresh:2147483647 Snd_cwnd:10 Advmss:1460 Reordering:3
	Rcv_rtt:0 Rcv_space:14600 Total_retrans:0 Pacing_rate:568810 Max_pacing_rate:18446744073709551615 Bytes_acked:1
	Bytes_received:0 Segs_out:2 Segs_in:1 Notsent_bytes:0 Min_rtt:50632 Data_segs_in:0 Data_segs_out:0 Delivery_rate:0
	Busy_time:0 Rwnd_limited:0 Sndbuf_limited:0 Delivered:1 Delivered_ce:0 Bytes_sent:0 Bytes_retrans:0 Dsack_dups:0
	Reord_seen:0 Rcv_ooopack:0 Snd_wnd:8192 Rcv_wnd:0 Rehash:0}
	*/
	fmt.Printf("%+v\n", tcpInfo)

	tcpConn.Write([]byte("hello"))
	// 获取 TCP 信息
	tcpInfo, err = unix.GetsockoptTCPInfo(fd, syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		log.Fatal(err)
	}
	/**
	&{State:1 Ca_state:0 Retransmits:0 Probes:0 Backoff:0 Options:6 Rto:248000 Ato:0 Snd_mss:1452 Rcv_mss:536 Unacked:1
	Sacked:0 Lost:0 Retrans:0 Fackets:0 Last_data_sent:0 Last_ack_sent:0 Last_data_recv:0 Last_ack_recv:0 Pmtu:1500
	Rcv_ssthresh:64076 Rtt:49093 Rttvar:24546 Snd_ssthresh:2147483647 Snd_cwnd:10 Advmss:1460 Reordering:3 Rcv_rtt:0
	Rcv_space:14600 Total_retrans:0 Pacing_rate:586641 Max_pacing_rate:18446744073709551615 Bytes_acked:1 Bytes_received:0
	Segs_out:3 Segs_in:1 Notsent_bytes:0 Min_rtt:49093 Data_segs_in:0 Data_segs_out:1 Delivery_rate:0 Busy_time:0
	Rwnd_limited:0 Sndbuf_limited:0 Delivered:1 Delivered_ce:0 Bytes_sent:5 Bytes_retrans:0 Dsack_dups:0 Reord_seen:0
	Rcv_ooopack:0 Snd_wnd:8192 Rcv_wnd:0 Rehash:0}
	*/
	fmt.Printf("%+v\n", tcpInfo)
}
