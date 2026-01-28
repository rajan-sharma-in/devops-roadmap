# Multi-stage build for small, non-root image
FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY app/go.mod ./
COPY app/*.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ops-demo

FROM gcr.io/distroless/static:nonroot
USER nonroot
WORKDIR /
COPY --from=builder /out/ops-demo /ops-demo
EXPOSE 8080
ENTRYPOINT ["/ops-demo"]
