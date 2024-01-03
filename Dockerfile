FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application
RUN go build -o myapp

# Expose a port the application will run on
EXPOSE 8080

# Command to run the executable
CMD ["./myapp"]