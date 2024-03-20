

# echo server
ncat -4lke $(which cat) 127.0.0.1 7777

# client
nc 127.0.0.1 7777
# Check the socket’s information
ss -4tlpne sport = 7777
#State   Recv-Q  Send-Q  Local Address:Port  Peer Address:Port  Process
#LISTEN  0       10      127.0.0.1:7777      0.0.0.0:*          users:(("ncat",pid=122701,fd=3)) uid:1000 ino:896838 sk:1 <->

