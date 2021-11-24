# syntax=docker/dockerfile:1

## Labels to apply to image
#LABEL author="Dr H"
#LABEL service="posts_service"
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

#*query dockerfile comparison
#Is WORKDIR / necessary?

COPY --from=build /service /service

EXPOSE 4000

#*query dockerfile comparison
#Is USER nonroot:nonroot necessary?

#*query dockerfile comparison
#CMD and ENTRYPOINT don't seem to be different in this case, are they?
CMD [ "/service" ]