############################################################ 
# Dockerfile to build golang Installed Containers 

# Based on alpine

############################################################

FROM tkeelio/go-proton-dev:v1

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn go build -o amqp-broker ./cmd/*
