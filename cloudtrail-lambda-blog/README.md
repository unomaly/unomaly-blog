# cloudtrail-lambda-unomaly-blog

This is a repo containing the necessary code and templates for getting up and
running with anomaly detection on your CloudTrail data using Unomaly!

## Deploying

**Requirements:** AWS SAM CLI installed.

Everything should be deployable using just the makefile, `template.yaml` defines
what should be created during the process and the `Makefile` is just used to
simplify things.

Essentially, you should just be able to go ahead and
```shell
# make deps
# make build
# make package
# make deploy
```
