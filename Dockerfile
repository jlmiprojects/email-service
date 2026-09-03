# Run Stage
FROM alpine:3.15.4

ARG version

# Set environment variable
ENV APP_NAME email-service

# Copy only required data into this image
COPY ./builds/$version/$APP_NAME .

VOLUME ["/templates"]


# Start app
CMD ./$APP_NAME
