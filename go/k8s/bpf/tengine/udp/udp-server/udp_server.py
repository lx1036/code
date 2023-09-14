

import socket

UDP_IP = "127.0.0.1"
UDP_PORT = 5000

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind((UDP_IP, UDP_PORT))

print("UDP server listening on {}:{}".format(UDP_IP, UDP_PORT))

while True:
    data, addr = sock.recvfrom(1024)
    print("Received message: {}, addr {}".format(data, addr))
    sock.sendto(b'Hello from the server!', addr)