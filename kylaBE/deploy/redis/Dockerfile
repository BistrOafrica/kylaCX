# Dockerfile for Redis deployment on AWS Fargate

FROM redis:8-alpine

# Redis password should be provided securely at runtime (e.g., via Docker secrets or environment variable)

# Set container hostname
ENV HOSTNAME=redis

# Expose Redis port
EXPOSE 6379

# Set up volume for data persistence
VOLUME ["/data"]

# Start Redis with password protection
CMD ["sh", "-c", "redis-server --requirepass \"${REDIS_PASSWORD:-38b9cf14dfbe29c51c666014040601e2}\""]
# Healthcheck for container
HEALTHCHECK --interval=5s --timeout=30s --retries=50 CMD ["sh", "-c", "redis-cli -a \"${REDIS_PASSWORD:-38b9cf14dfbe29c51c666014040601e2}\" ping || exit 1"]% 