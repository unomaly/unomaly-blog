.PHONY: deps clean build

deps:
	dep ensure
#	go get -u ./...

clean: 
	rm -rf ./cloudtrailToUnomaly/* ./build/*
	
build:
	GOOS=linux GOARCH=amd64 go build -o build/cloudtrail-lambda-unomaly-blog ./cloudtrailToUnomaly/main.go

package:
	sam package \
		--template-file template.yaml \
		--output-template-file packaged.yaml \
		--s3-bucket unomaly-lambda-functions

deploy:
	sam deploy \
		--template-file packaged.yaml \
		--stack-name cloudtrail-lambda-unomaly-blog \
		--capabilities CAPABILITY_IAM