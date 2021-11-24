# syntax=docker/dockerfile:1

## Labels to apply to image
#LABEL author="Dr H"
#LABEL service="comments_service"
#LABEL version="1.0"

# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.17-bullseye as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

# Download necessary Go modules
RUN go mod download

COPY *.go ./

RUN go build -o /service


FROM gcr.io/distroless/base-debian11

COPY --from=build /service /service

EXPOSE 4001

CMD [ "/service" ]