
import http.server
import socket

class ProxyProtocolHandler(http.server.SimpleHTTPRequestHandler):
    def handle_one_request(self):
        # 解析并验证Proxy Protocol报文
        proxy_protocol_header = self.parse_proxy_protocol_header()
        if not proxy_protocol_header:
            self.send_error(400, 'Invalid Proxy Protocol header')
            return

        # 获取真实的客户端地址
        client_address = (proxy_protocol_header['ip'], proxy_protocol_header['port'])

        # 设置socket的客户端地址为真实的客户端地址
        self.request.setpeername(client_address)

        # 调用父类的handle_one_request方法处理请求
        super().handle_one_request()

    def parse_proxy_protocol_header(self):
        # 从请求socket中读取Proxy Protocol报文
        header_data = self.request.recv(108)

        # 解析Proxy Protocol报文
        # 报文格式：PROXY <protocol> <client IP> <proxy IP> <client port> <proxy port> <\r\n>
        header_parts = header_data.decode('utf-8').strip().split('\r\n')
        header_parts = header_parts[0].strip().split(' ')
        if len(header_parts) != 6 or header_parts[0] != 'PROXY':
            return None

        proxy_protocol_header = {
            'protocol': header_parts[1],
            'src_ip': header_parts[2],
            'src_port': int(header_parts[4])
        }

        return proxy_protocol_header

# 创建HTTP服务器，并使用自定义的请求处理程序
server_address = ('', 12345)
httpd = http.server.HTTPServer(server_address, ProxyProtocolHandler)

# 启动HTTP服务器
httpd.serve_forever()


# python3 proxy_protocol_server.py

