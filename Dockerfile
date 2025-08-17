# Multi-stage Dockerfile for production

FROM node:22-alpine AS builder
WORKDIR /app

# Install dependencies
COPY package.json ./
# If you add a lockfile later, copy it as well for reproducible builds
# COPY yarn.lock ./
RUN corepack enable && (yarn install --frozen-lockfile || yarn install)

# Copy sources and build
COPY tsconfig.json ./
COPY src ./src
RUN yarn build


FROM node:22-alpine AS runner
ENV NODE_ENV=production
WORKDIR /app

# Install only production dependencies
COPY package.json ./
# COPY yarn.lock ./
RUN corepack enable && (yarn install --production=true --frozen-lockfile || yarn install --production=true)

# Copy built artifacts
COPY --from=builder /app/dist ./dist
COPY scripts ./scripts
COPY drizzle/migrations ./drizzle/migrations
COPY docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["sh", "/app/docker-entrypoint.sh"]


