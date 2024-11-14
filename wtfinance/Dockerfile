# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.22.2-alpine AS builder

WORKDIR /code

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download


ENV SERVER_PORT=80
ENV SECRETS_PATH=/run/secrets/googleSheets

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/cache \
  --mount=type=cache,target=/go-build \
  go build -o bin/wtfinance main.go

CMD ["/code/bin/wtfinance"]

FROM builder as dev-envs

RUN <<EOF
apk update
apk add git
EOF

RUN <<EOF
addgroup -S docker
adduser -S --shell /bin/bash --ingroup docker vscode
EOF

# install Docker tools (cli, buildx, compose)
COPY --from=gloursdocker/docker / /

CMD ["go", "run", "main.go"]

FROM scratch
COPY --from=builder /code/bin/wtfinance /usr/local/bin/wtfinance
CMD ["/usr/local/bin/wtfinance"]