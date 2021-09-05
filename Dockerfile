#build stage
FROM golang:1.17.0-alpine3.14 AS build-env
ARG GH_TOKEN
RUN apk add build-base git
RUN git config --global url."https://${GH_TOKEN}:x-oauth-basic@github.com/ProjectAthenaa".insteadOf "https://github.com/ProjectAthenaa"
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -o client_proxy_service


# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /app/client_proxy_service /app/
EXPOSE 3000 3000
ENTRYPOINT ./client_proxy_service