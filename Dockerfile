FROM golang:1.14-alpine AS go-build

ARG GITLAB_TOKEN
ARG GITLAB_DOMAIN

RUN mkdir /src
WORKDIR /src

# Git installs
RUN apk add --update --no-cache ca-certificates git

RUN git config --global url."https://root:$GITLAB_TOKEN@$GITLAB_DOMAIN".insteadOf "https://$GITLAB_DOMAIN"

ENV GO111MODULE=on
ENV GOPRIVATE=gl.ocelotworks.com/*
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOPATH=/src/go

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine

RUN apk --no-cache --update add ca-certificates curl
WORKDIR /app
COPY --from=go-build /src/main /app/
RUN mkdir /app/res
COPY --from=go-build /src/res/ /app/res/
RUN mkdir /app/output
COPY crontab.txt crontab.txt
RUN crontab crontab.txt
EXPOSE 2112
HEALTHCHECK --interval=2m --start-period=1m --retries=5 \
    CMD curl -f http://localhost:2112/healthz || exit 1
ENTRYPOINT crond && ./main