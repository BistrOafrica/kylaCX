# Use official Envoy image as base
FROM envoyproxy/envoy:v1.28-latest

# Copy custom Envoy configuration into the container
COPY deploy/envoy/envoy.yaml /etc/envoy/envoy.yaml

# Set the default command to run Envoy with the custom config
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml", "--service-node", "local-envoy", "--service-cluster", "local-cluster"]