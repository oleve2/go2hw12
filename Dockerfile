FROM alpine:3.7
RUN mkdir /app
ADD server_new /app/
ENTRYPOINT ["/app/server_new"]
EXPOSE 9999

