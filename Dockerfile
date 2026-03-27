FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gate ./cmd/gate

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /gate /gate
COPY migrations/ /migrations/
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/gate"]
