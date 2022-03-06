AMQP_VERSION := 0.4.0-alpha.3 
docker-build:
	sudo docker build -t tkeelio/amqp:${AMQP_VERSION} .
docker-push:
	sudo docker push tkeelio/amqp:${AMQP_VERSION}

docker-auto:
	sudo docker build -t tkeelio/amqp:${AMQP_VERSION} .
	sudo docker push tkeelio/amqp:${AMQP_VERSION}