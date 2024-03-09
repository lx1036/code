

# ecs 升级内核

```shell
sudo apt update -y
apt-cache search linux-image
apt-get install -y linux-image-6.5.0-14-generic
sudo update-grub
sudo reboot
uname -r
```
