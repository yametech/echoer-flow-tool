all: build-base build-controller


build-base:
	docker build -t harbor.ym/devops/verthandi-base:v0.0.2 -f docker/Dockerfile.base .
	docker push harbor.ym/devops/verthandi-base:v0.0.2

build-controller:
	docker build -t harbor.ym/devops/verthandi-controller:v0.0.2 -f docker/Dockerfile.pipeline-controller .
	docker push harbor.ym/devops/verthandi-controller:v0.0.2
