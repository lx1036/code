


import socket
import ssl
import os

HOST = 'localhost'
PORT = 5005

# 创建基础 TCP 套接字
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

# 创建 SSL 套接字包装基础套接字
ca_pem = os.path.abspath("{}/../../ssl/ca.pem".format(os.path.dirname(__file__)))
ssl_sock = ssl.wrap_socket(sock, ca_certs=ca_pem)

# 连接服务器
ssl_sock.connect((HOST, PORT))

# 发送数据
ssl_sock.sendall(b'GET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n')

# 接收响应数据
response = b''
while True:
    data = ssl_sock.recv(4096)
    if not data:
        break
    response += data

# 打印响应数据
print(response.decode())

# 关闭连接
ssl_sock.close()