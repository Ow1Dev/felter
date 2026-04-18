# Builder-only Dockerfile: builds Angular app artifacts with Bun + Node
# No runtime server; resulting image contains /app/dist for CI/artifacts

FROM node:20-alpine AS build
WORKDIR /app

# Install Bun
RUN apk add --no-cache curl ca-certificates \
  && curl -fsSL https://bun.sh/install | bash \
  && mv /root/.bun/bin/bun /usr/local/bin/bun

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npx @angular/cli@21.2.7 version >/dev/null 2>&1 || true \
  && npx ng build --configuration production

# Artifact image containing dist only (no entrypoint)
FROM scratch AS artifact
COPY --from=build /app/dist /dist
