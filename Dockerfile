FROM registry.redhat.io/ubi9:9.0.0-1576
WORKDIR /app
COPY . .
CMD ["/app/fx-test"]
