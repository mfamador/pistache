# Use the latest alpine image for the base.
FROM alpine:latest

# Install bash and the iptables command.
RUN apk update && apk add --no-cache bash iptables

# Copy the initialization script into the container.
COPY build/init/init.sh /bin/

# Mark the initialization script as executable.
RUN chmod +x /bin/init.sh

# Start the entrypoint script on container startup.
ENTRYPOINT ["init.sh"]
