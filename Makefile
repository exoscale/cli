

docker-build:
	docker build -t exo --build-arg VCS_REF="$(shell git rev-parse HEAD)" \
	--build-arg BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" . 
	docker tag exo $(DOCKER_ID_USER)/exo
	docker push $(DOCKER_ID_USER)/exo