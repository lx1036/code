


# 需要内核 >6.0，没有测试成功
config_device()
{
  ip link add eth-svc type dummy
  ip link add eth-ep1 type dummy
  ip link add eth-ep2 type dummy
  ip link set eth-svc up
  ip link set eth-ep1 up
  ip link set eth-ep2 up
  ip addr add 10.240.2.1/24 dev eth-svc
  ip addr add 10.240.1.2/24 dev eth-ep1
  ip addr add 10.240.1.3/24 dev eth-ep2
}

config_device
