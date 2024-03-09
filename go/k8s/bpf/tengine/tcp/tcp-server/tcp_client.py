import socket
import time

# 服务器的 IP 地址和端口号
host = "127.0.0.1"
# port = 8081
port = 6000

start_time = time.time()


# 创建 TCP 客户端套接字
client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
# client_socket.setblocking(False)

# try:

# 连接到服务器
client_socket.connect((host, port))
print(f"已连接到服务器 {host}:{port}")

# for i in range(10):
# 发送数据到服务器
# 不加 "Connection: close" 短连接，默认为 "Connection: keep-alive" 为长连接，连接等 keep-alive timeout 后才会断掉
message = "GET / HTTP/1.1\nHost: {}\nUser-Agent: curl/7.89.0\nConnection: close\nAccept: */*\n\n".format(host)
# message = "hello"
# message = "GET / HTTP/1.1\r\nHost: {}\r\nUser-Agent: curl/7.87.0\r\nConnection: close\r\nAccept: */*\r\n\r\n".format(host)
client_socket.send(message.encode())
# client_socket.settimeout(3)
# 接收数据的缓冲区大小
# 1024, 1024 取值没有 <h1>hello world</h1> 值，所以这里需要处理 recv(bytes) 多少字节数量
buffer_size = 1024
all_data = b""
try:
    # 接收服务器的响应数据
    while True:
        response = client_socket.recv(buffer_size)
        if len(response) == 0:
            print(response) # nginx 的时候，会发一个空包, response = b''
            # client_socket.close()
            print("close client_socket")
            break
        # if response:
            # print(f"收到服务器的响应: {response.decode()}")
        # print(f"收到服务器的响应: {response.decode()}")
        all_data += response
        # print(f"临时的收到服务器的响应: {all_data.decode()}")
        # else:
        #     client_socket.close()
        #     break
        # 如果接收到的数据为空，则表示服务器关闭连接
        # if not response:
        #     break
        

    print(f"收到服务器的响应: \n\n{all_data.decode()}")
except socket.error as e:
    print(f"与服务器 {host}:{port} 的连接发生错误: {e}")

finally:
    # 关闭客户端套接字
    client_socket.close()

# 云监控 TCP 拨测肯定没有 recv 服务端数据，直接发包过去，然后 close
# client_socket.close()

end_time = time.time()
run_time = end_time - start_time
print("代码运行时间为：", run_time * 1000, "ms")