FROM golang:1.17
RUN apt-get -y update
    # Required dependencies
RUN apt-get install gcc g++ cmake cmake-curses-gui uuid-dev -y

    # Dependencies needed for SSL support
RUN apt-get install libssl-dev -y

    # dependencies needed for Cyrus SASL support
RUN apt-get install libsasl2-2 libsasl2-dev libsasl2-modules -y

RUN wget https://github.com/apache/qpid-proton/archive/refs/tags/0.36.0.tar.gz
RUN tar xvf 0.36.0.tar.gz
RUN mkdir /go/qpid-proton-0.36.0/build
RUN cd /go/qpid-proton-0.36.0/build && cmake .. -DCMAKE_INSTALL_PREFIX=/usr -DSYSINSTALL_BINDINGS=OFF && make install
