#https://github.com/slok/kube-code-generator/blob/master/Dockerfile
FROM golang:1.13.8
ARG CODEGEN_VERSION="1.17.3"

# Code generator stuff
RUN wget http://github.com/kubernetes/code-generator/archive/kubernetes-${CODEGEN_VERSION}.tar.gz && \
    mkdir -p ${CODE_GENERATOR_PATH} && \
    tar zxvf kubernetes-${CODEGEN_VERSION}.tar.gz --strip 1 -C ${CODE_GENERATOR_PATH} && \
    rm kubernetes-${CODEGEN_VERSION}.tar.gz

CMD ["./update-codegen.sh"]
