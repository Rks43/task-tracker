FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./
COPY healthcheck.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o task-tracker main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o healthcheck healthcheck.go

FROM scratch

COPY --from=builder /app/task-tracker /task-tracker
COPY --from=builder /app/healthcheck /healthcheck

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD ["/healthcheck"]

ENTRYPOINT ["/task-tracker"]
