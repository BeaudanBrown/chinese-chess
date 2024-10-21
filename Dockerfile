FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
RUN mkdir -p ./data
COPY ./www ./www
COPY ./dist/main .
EXPOSE 8069
CMD ["./main"]
