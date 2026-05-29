# Frontend image: builds the Vite SPA, serves it with nginx, and reverse-
# proxies /grpc → envoy:8081 so the SPA can hit the gRPC-web gateway
# same-origin.
#
# Build context: monorepo root (kylaCX/).

FROM node:22-slim AS builder
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.9 --activate

COPY kylaFE/package.json kylaFE/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY kylaFE/ ./

# Build args are forwarded as VITE_* env vars so Vite picks them up at build.
ARG VITE_API_URL=/grpc
ARG VITE_SENTRY_DSN=
ARG VITE_ENVIRONMENT=local
ENV VITE_API_URL=${VITE_API_URL} \
    VITE_SENTRY_DSN=${VITE_SENTRY_DSN} \
    VITE_ENVIRONMENT=${VITE_ENVIRONMENT}

RUN pnpm build

FROM nginx:1.27-alpine AS production
COPY kylaOPs/docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
