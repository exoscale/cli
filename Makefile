VCS_REF="$(shell git rev-parse HEAD)"

BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%m:%SZ")" 

docker-build:
	docker build -t exo .
	docker tag exo $(DOCKER_ID_USER)/exo
	docker push $(DOCKER_ID_USER)/exo