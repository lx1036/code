



import socket


# tcpdump -i eth0 -vveenn -A port 8890 -w tcp_fastopen_8890.pcap

def handle_connection(conn):
    data = conn.recv(1024)
    print("Received message:", data.decode())

    # 回复客户端
    conn.sendall(b"Hello, client!")
    conn.close()

def main():
    server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server.setsockopt(socket.SOL_TCP, socket.TCP_FASTOPEN, 10)
    server.bind(('0.0.0.0', 8890))
    server.listen(5)

    while True:
        conn, addr = server.accept()
        handle_connection(conn)

if __name__ == "__main__":
    main()

