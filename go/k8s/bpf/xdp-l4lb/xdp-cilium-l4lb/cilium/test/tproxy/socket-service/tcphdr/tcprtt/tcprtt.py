

import socket
import struct

def get_tcp_info(sock):
    fmt = "B"*7+"I"*21
    # x = sock.getsockopt(socket.IPPROTO_TCP, socket.TCP_INFO, 48)
    x = sock.getsockopt(socket.IPPROTO_TCP, socket.TCP_INFO, 92)
    info = struct.unpack(fmt, x)
    print(info)

def main():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    get_tcp_info(s)
    print("----------\n")

    s.connect(('www.baidu.com', 80))
    get_tcp_info(s)
    print("----------\n")

    s.send(b"hi\n\n")
    get_tcp_info(s)
    print("----------\n")

    s.recv(1024)
    get_tcp_info(s)

# 只有 linux 才有 socket option socket.TCP_INFO，必须 linux 中运行: python3 tcprtt.py

# (7, 0, 0, 0, 0, 0, 0, 1000000, 0, 536, 0, 0, 0, 0, 0, 0, 10403928, 0, 10403928, 10403928, 0, 0, 0, 250000, 2147483647, 10, 0, 3)
# ----------

# (1, 0, 0, 0, 0, 6, 117, 204000, 0, 1452, 536, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1500, 64076, 2114, 1057, 2147483647, 10, 1460, 3)
# ----------

# (1, 0, 0, 0, 0, 6, 117, 204000, 0, 1452, 536, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1500, 64076, 2114, 1057, 2147483647, 10, 1460, 3)
# ----------

# (1, 0, 0, 0, 0, 6, 117, 204000, 40000, 1452, 536, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1500, 64076, 2132, 829, 2147483647, 10, 1460, 3)
if __name__ == "__main__":
    main()



