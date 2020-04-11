# build the go server
FROM golang:1.13-alpine AS server_builder
ADD . /app
WORKDIR /app/server
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -a -o /main .

# build the react static assets
FROM node:alpine3.10 AS node_builder
RUN apk --no-cache add build-base python2
COPY --from=server_builder /app/client ./
RUN npm ci --production
RUN npm run build

# production container with all dependencies stripped
FROM alpine:3.11
RUN apk --no-cache add ca-certificates
COPY --from=server_builder /main ./
COPY --from=server_builder /app/server/static ./static
COPY --from=node_builder /build ./static/web

# run executable go server
RUN chmod +x ./main
ENTRYPOINT ./main
