

import socket

UDP_IP = "127.0.0.1"
UDP_PORT = 5000

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
# sock.setsockopt(socket.IPPROTO_IP, socket.IP_TRANSPARENT, 1)
sock.bind((UDP_IP, UDP_PORT))

print("UDP server listening on {}:{}".format(UDP_IP, UDP_PORT))

while True:
    data, client_address = sock.recvfrom(1024)
    print("Received message: {}, addr {}".format(data, client_address))
    response = "Hello from the server!"
    # sock.sendto(b'Hello from the server!', client_address)
    sock.sendto(response.encode(), client_address)
    # sock.send(response.encode())