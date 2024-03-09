
#include <stdio.h>
typedef unsigned int __u32;


static char* u32toIpStr(__u32 ip) {
    static char str[16]; // IP地址字符串长度最大为"255.255.255.255"即15个字符加结束符'\0'

    // 分离IP地址的四个字节
    unsigned char bytes[4];
    bytes[0] = (ip >> 24) & 0xFF;
    bytes[1] = (ip >> 16) & 0xFF;
    bytes[2] = (ip >> 8) & 0xFF;
    bytes[3] = ip & 0xFF;

    // 将字节转换为点分十进制字符串
    sprintf(str, "%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3]);

    return str;
}

int main() {
    char* ip = u32toIpStr(0xad100164);
    printf("%s\n", ip); // 173.16.1.100
}
