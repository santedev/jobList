FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Download Tailwind CSS binary
RUN curl -sLo tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
RUN chmod +x tailwindcss

# Copy the source code into the container
COPY . .

# Compile Tailwind CSS
RUN ./tailwindcss -i tailwind/css/app.css -o public/styles.css

# Clean up Tailwind CSS CLI tool
RUN rm tailwindcss

# Generate templ files
RUN /go/bin/templ generate

# Build the Go application
RUN go build -o bin/app .

# Command to run the application
CMD ["./bin/app"]