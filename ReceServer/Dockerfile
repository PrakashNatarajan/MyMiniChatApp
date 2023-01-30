FROM golang:1.19-alpine

# Set the Current Working Directory inside the container
WORKDIR /app/receserver

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/receserver .


# This container exposes port 5600 to the outside world
EXPOSE 5600

# Run the binary program produced by `go install`
CMD ["./out/receserver"]
