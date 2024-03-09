

typedef unsigned int __u32;

struct tcpbpf_globals {
    __u32 event_map;
};

int main() {
    struct tcpbpf_globals g = {};
//    g = {};

    return 0;
}
