FROM alpine:3.21.3

ARG TARGETARCH

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY bin/linux_${TARGETARCH}/stackdome-server .

USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["./stackdome-server"]
CMD ["serve"]
