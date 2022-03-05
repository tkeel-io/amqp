############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM tkeelio/go-proton-dev:v1 as builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn go build -o amqp-broker ./cmd/*

FROM tkeelio/go-proton-dev:v1

RUN mkdir /keel
COPY --from=builder /src/amqp-broker /keel

EXPOSE 5672
WORKDIR /keel
CMD ["/keel/amqp-broker"]