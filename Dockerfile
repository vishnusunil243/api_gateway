#Build stage
FROM golang:1.21-bullseye AS builder

RUN apt-get update

WORKDIR /apigateway

COPY . .
 
COPY .env .env

RUN go mod download

RUN go build -o ./out/dist ./cmd

#production stage

FROM busybox

RUN mkdir -p /apigateway/out/dist

COPY --from=builder /apigateway/out/dist /apigateway/out/dist

COPY --from=builder /apigateway/.env /apigateway/out

WORKDIR /apigateway/out/dist

EXPOSE 8081

CMD ["./dist"]