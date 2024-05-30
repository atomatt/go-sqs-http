ARG GOVERSION=1.22

FROM golang:${GOVERSION} AS build
WORKDIR /code
COPY go.mod go.sum ./
RUN go mod download
COPY ./* ./
RUN CGO_ENABLED=0 GOOS=linux go build -v .

FROM scratch
COPY --from=build /code/demo /
ENTRYPOINT [ "/demo" ]
