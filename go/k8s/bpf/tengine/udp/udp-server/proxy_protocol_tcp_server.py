
import socket

def parse_proxy_protocol_header(client_socket):
    # 从客户端socket中读取Proxy Protocol报文
    header_data = client_socket.recv(108)
    print(header_data)

    # 解析Proxy Protocol报文
    # 报文格式：PROXY <protocol> <client IP> <proxy IP> <client port> <proxy port> <\r\n>
    header_parts = header_data.decode('utf-8').strip().split('\r\n')
    header_parts = header_parts[0].strip().split(' ')
    # print(header_parts)
    if len(header_parts) != 6 or header_parts[0] != 'PROXY':
        return None

    proxy_protocol_header = {
        'protocol': header_parts[1],
        'src_ip': header_parts[2], # src_ip
        'src_port': int(header_parts[4]) # src_port
    }

    return proxy_protocol_header

def handle_client(client_socket):
    # 解析并验证Proxy Protocol报文
    proxy_protocol_header = parse_proxy_protocol_header(client_socket)
    if not proxy_protocol_header:
        client_socket.sendall(b'HTTP/1.1 400 Bad Request\r\n\r\nInvalid Proxy Protocol header')
        client_socket.close()
        return
    
    print(proxy_protocol_header)
    # 获取真实的客户端地址
    client_address = (proxy_protocol_header['src_ip'], proxy_protocol_header['src_port'])
    # 设置socket的客户端地址为真实的客户端地址
    # client_socket.setsockopt(socket.SOL_SOCKET, socket.SO_PEERADDR, client_address) # linux
    # 处理请求
    msg = "HTTP/1.1 200 OK\r\n\r\nWelcome {}:{}".format(proxy_protocol_header['src_ip'], proxy_protocol_header['src_port'])
    client_socket.sendall(msg.encode())
    client_socket.close()

# 创建TCP服务器
server_address = ('', 12345)
server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server_socket.bind(server_address)
server_socket.listen(5)

print('Server listening on {}:{}'.format(*server_address))

while True:
    # 等待客户端连接
    client_socket, client_address = server_socket.accept()
    print('Received connection from {}:{}'.format(*client_address))

    # 处理客户端连接
    handle_client(client_socket)
    print("\n")



# python3 proxy_protocol_server.py

