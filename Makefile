build:
	docker build --build-arg GH_TOKEN=$(token)  -t registry.digitalocean.com/athenabot/general/proxy-client-service:latest .
