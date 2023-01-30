FROM golang:1.19-alpine

# Set the Current Working Directory inside the container
WORKDIR /app/sendserver

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/sendserver .


# This container exposes port 6500 to the outside world
EXPOSE 6500

# Run the binary program produced by `go install`
CMD ["./out/sendserver"]
