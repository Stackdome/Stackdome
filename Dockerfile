FROM alpine:3.21.3

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY bin/linux_amd64/stackdome-server .

USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["./stackdome-server"]
CMD ["serve"]
