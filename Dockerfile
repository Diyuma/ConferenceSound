FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .
# TODO later copy not all

# Build the Go application
RUN go build -C server/cmd -o server

# Expose a port the application will run on
EXPOSE 9090

# Command to run the executable
CMD ["./server/cmd/server", "-sredisaddr", "172.18.0.5:6379"]