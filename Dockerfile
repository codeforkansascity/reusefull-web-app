FROM debian:10-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY main /
COPY templates /templates
COPY assets /assets
ENTRYPOINT ["./main"]
