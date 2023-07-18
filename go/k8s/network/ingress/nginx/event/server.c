

// g++ -o server server.c // 还不能是 cc，否则找不到头文件
// ./server
// https://levelup.gitconnected.com/nginx-event-driven-architecture-demonstrated-in-code-51bf0061cad9
// https://www.cnblogs.com/xuewangkai/p/11158576.html


#include <cerrno>
#include <unistd.h>
#include <stdio.h>
#include <sys/epoll.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <signal.h>
#include <stdlib.h>
#include <netinet/in.h>
#include <fcntl.h>
#include <string.h>

#define PORT 6868
#define MAX_CONNS 100
#define MAX_EVENTS 1024
#define MAX_BUFFER_LEN 1024
#define WORKER 4

static volatile bool running = true;

void serve(pid_t pid, int sfd, struct sockaddr* addr, socklen_t* len);

void abort(const char msg[]) {
  perror(msg);
  exit(1);
}

void setNonblocking(int fd) {
  int flags = fcntl(fd, F_GETFL, 0);
  if (flags == -1) {
    abort("Failed to obtain FD flags");
  }
  flags |= SOCK_NONBLOCK;
  if (fcntl(fd, F_SETFL, flags) < 0) {
    abort("Failed to set FD to nonblocking");
  }
}

void registerFd(int efd, int fd) {
  struct epoll_event event;
  event.data.fd = fd;
  event.events = EPOLLIN; // 表示对应的文件描述符可以读（包括对端SOCKET正常关闭）
//  EPOLLOUT：表示对应的文件描述符可以写；
// EPOLLPRI：表示对应的文件描述符有紧急的数据可读（这里应该表示有带外数据到来）；
// EPOLLERR：表示对应的文件描述符发生错误；
// EPOLLHUP：表示对应的文件描述符被挂断；
// EPOLLET： 将EPOLL设为边缘触发(Edge Triggered)模式，这是相对于水平触发(Level Triggered)来说的。
// EPOLLONESHOT：只监听一次事件，当监听完这次事件之后，如果还需要继续监听这个socket的话，需要再次把这个socket加入到EPOLL队列里

    // 2.
    if (epoll_ctl(efd, EPOLL_CTL_ADD, fd, &event) < 0) {
        abort("Failed to add FD to epoll");
    }
}

void shutdown(int) {
  running = false;
}

// 使用 epoll 让 linux 来 monitor 一系列 fd，如果 fd 有 read/write 事件，告知当前 server 进程
void serve(pid_t pid, int sfd, struct sockaddr* addr, socklen_t* len) {
    // 1.
    int efd = epoll_create1(0);
    if (efd < 0) {
        abort("Failed to create epoll event");
    }

    // 把当前 socket fd 注册到 epoll 中，内核会监听该 socket fd
    // socket fd 可以读!!!
    registerFd(efd, sfd);

    struct epoll_event* events = (struct epoll_event*)calloc(MAX_EVENTS, sizeof(struct epoll_event));
    while(running) {
        // 3. 阻塞等待
        int n = epoll_wait(efd,
        events, /* Empty events that will be filled on return */ // events则是分配好的 epoll_event结构体数组，epoll将会把发生的事件复制到 events数组中
        MAX_EVENTS,
        1000/* timeout in ms */);
        // printf("%d epoll events is produced...\n", n);
        for (int i = 0 ; i < n; i ++) {
            // Process events[i]. It has IO updates.

            // Error check.
            if (events[i].events & EPOLLERR || events[i].events & EPOLLHUP || !(events[i].events & EPOLLIN)) {
                // Error occurred. Ignore.
                continue;
            }
            if (events[i].data.fd == sfd) { //有新的连接
                // Events on listening socket.
                int cfd = accept(sfd, addr, len); // client socket fd, accept 这个连接
                if (cfd == -1 && (errno == EAGAIN || errno == EWOULDBLOCK)) {
                    // Empty event on the listening queue. Other processes have grabbed it.
                    continue;
                } else if (cfd < 0) {
                    abort("Failed to accept client socket");
                }

                setNonblocking(cfd);
                registerFd(efd, cfd); // 将新的fd添加到epoll的监听队列中
                printf("register accept fd\n");
            }

            /*

                else if( events[i].events&EPOLLIN ) //接收到数据，读socket
                {
                    n = read(sockfd, line, MAXLINE)) < 0    //读
                    ev.data.ptr = md;     //md为自定义类型，添加数据
                    ev.events=EPOLLOUT|EPOLLET;
                    epoll_ctl(epfd,EPOLL_CTL_MOD,sockfd,&ev);//修改标识符，等待下一个循环时发送数据，异步处理的精髓
                }
                else if(events[i].events&EPOLLOUT) //有数据待发送，写socket
                {
                    struct myepoll_data* md = (myepoll_data*)events[i].data.ptr;    //取数据
                    sockfd = md->fd;
                    send( sockfd, md->ptr, strlen((char*)md->ptr), 0 );        //发送数据
                    ev.data.fd=sockfd;
                    ev.events=EPOLLIN|EPOLLET;
                    epoll_ctl(epfd,EPOLL_CTL_MOD,sockfd,&ev); //修改标识符，等待下一个循环时接收数据
                }

            */


            else {
                // Event on read.
                char buffer[512];
                int l = read(events[i].data.fd, buffer, MAX_BUFFER_LEN);
                buffer[l] = 0;
                printf("Pid: %d => %s\n", pid, buffer);
                // Done with the client socket. Close it to remove it from the epoll list.
                close(events[i].data.fd);
            }
        }
    }
    free(events);
    close(efd);
}

int main(void) {
    signal(SIGINT, shutdown);

    struct sockaddr_in address;
    int addrlen = sizeof(address);
    int sfd;
    // create socket fd
    sfd = socket(AF_INET, SOCK_STREAM | SOCK_NONBLOCK, 0);
    if (sfd == 0) {
        abort("failed to create a server socket");
    }
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(PORT);
    if (bind(sfd, (struct sockaddr *)&address, sizeof(address)) < 0) {
        abort("failed to bind a server socket");
    }
    if (listen(sfd, MAX_CONNS) < 0) {
        abort("Failed to listen to server socket");
    }

    printf("Server started...\n");
    for (int i = 0; i < WORKER; i ++) {
        if (fork() == 0) {
            // Child process.
            pid_t pid = getpid();
            printf("Worker %d is ready\n", pid);
            serve(pid, sfd, (struct sockaddr *)&address, (socklen_t*)&addrlen);
            printf("Worker %d shuts down\n", pid);
            return 0;
        }
    }

    while(running) {
        sleep(1);
    }
    close(sfd);
    printf("Gracefully exiting the program\n");
    return 0;
}
