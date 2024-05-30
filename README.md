# Example Go process with JSON HTTP API and SQS consumer loop

Highlights:

* One Go process
* Super simple JSON HTTP API implementation
* SQS consumer loop, with per-message handler
* Backoff for the consumer loop
* Structured logs
* Graceful shutdown, with limit
* Docker Compose and LocalStack dev setup
* AWS setup with Terraform
* S3 bucket -> SQS notification to poke the consumer

`docker compose up` should be enough to get it going.

Trigger a SQS message by creating a new S3 file, e.g.
`N=a; echo $N | docker-compose run --rm -T localstack-cli s3 cp - s3://outbox/$N`.
