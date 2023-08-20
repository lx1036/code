
import socket

def main():
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    client.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    client.setsockopt(socket.SOL_TCP, socket.TCP_FASTOPEN, 10)
    addr = ('127.0.0.1', 8890)
    # client.connect(addr) # sendto 已经包含 connect 去连接 server

    msg = "Hello, server!"
    # client.sendall(msg.encode(), socket.MSG_FASTOPEN) # 参数有问题
    client.sendto(msg.encode(), socket.MSG_FASTOPEN, addr) # socket.MSG_FASTOPEN 只有 linux 有

    data = client.recv(1024)
    print("Received response:", data.decode())

    client.close()

if __name__ == "__main__":
    main()