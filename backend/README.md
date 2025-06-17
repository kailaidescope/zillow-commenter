# Readme for Backend work

### Lambda Setup

Use [AWS' guide](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html) on how to package Go executables for Lambda.

Use [Lambda env variables](https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html) as you would for local Go env variables.

Use AWS Gateway to [proxy](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html) to Lambda.

## Go commands:

### Mac/Linux

```
GOOS=linux
GOARCH=arm64
CGO_ENABLED=0
go build -tags lambda.norpc -o ./bin/x.x.x/bootstrap main.go
zip ./bin/x.x.x/zillowette_lambda.zip ./bin/x.x.x/bootstrap
```

### Windows

```
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
go build -tags lambda.norpc -o ./bin/bootstrap main.go
~\Go\Bin\build-lambda-zip.exe -o ./bin/zillowette_lambda.zip ./bin/bootstrap
```

AWS CLI commands:

```
aws lambda create-function --function-name zillowette --runtime provided.al2023 --handler bootstrap --architectures arm64 --role arn:aws:iam::111122223333:role/lambda-exec --zip-file fileb://bin/zillowette_lambda.zip
```
