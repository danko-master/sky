FROM golang:1.25

# Set the working directory inside the container
WORKDIR /app

# Copy your Go application source code into the container
COPY . .

# Build your Go application
RUN go build -o myapp .

# Define the command to run your application
CMD ["./myapp"]