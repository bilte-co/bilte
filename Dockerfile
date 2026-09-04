FROM golang:1.27-trixie AS base
WORKDIR /app
COPY . .
RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "-w -s" -o ./bilte ./cmd

# Static binary (CGO_ENABLED=0), so the libc-free static image is enough.
FROM gcr.io/distroless/static-debian13

WORKDIR /app

# The server reads ./data and ./static from disk at runtime.
COPY --from=base /app /app

ENTRYPOINT ["/app/bilte"]
CMD ["web"]
