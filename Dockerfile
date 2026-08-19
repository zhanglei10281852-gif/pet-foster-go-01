FROM golang:1.22 AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/pet-foster-server ./cmd/pet-server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/pet-foster-server /pet-foster-server
EXPOSE 8080
ENTRYPOINT ["/pet-foster-server"]
