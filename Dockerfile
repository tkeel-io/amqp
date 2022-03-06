############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM tkeelio/go-proton-dev:v2 as builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn go build -o amqp-broker ./cmd/*

FROM tkeelio/proton-runtime:v2

RUN mkdir /keel
COPY --from=builder /src/amqp-broker /keel

EXPOSE 5672
WORKDIR /keel
CMD ["/keel/amqp-broker"]
