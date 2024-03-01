
import socket
import os

# 创建一个TCP/IP socket
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
# 绑定到所有接口的特定端口
server_address = ('127.0.0.1', 10000)
sock.bind(server_address)
# 开始监听连接
sock.listen(1)

# 获取当前进程的PID
current_pid = os.getpid()
print("当前进程的PID:", current_pid)

while True:
    # 等待连接
    connection, client_address = sock.accept()
    print('连接来自', client_address)

    # client_address 是一个元组，包含 IP 和端口
    client_ip, client_port = client_address
    print('客户端IP:', client_ip)
    print('客户端端口:', client_port)

    # 处理客户端连接...
    # 关闭连接
    connection.close()

# python3 client-port.py
# tubectl bind foo tcp 127.0.0.1 4322
# tubectl register-pid 485850 foo tcp 127.0.0.1 10000
# echo hello | nc -q 1 127.0.0.1 4322

# 保留源目的端口 4322，而不是 10000，验证通过!!!
# 抓包 4322 有包: tcpdump -i lo port 4322
# 抓包 10000 没有包: tcpdump -i lo port 10000

