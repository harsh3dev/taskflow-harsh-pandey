# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:20-alpine AS build

WORKDIR /app

# Copy manifests first for layer caching.
COPY package*.json ./

# Use `npm install` (not `npm ci`) so npm resolves the correct platform-specific
# native binaries for the build environment (linux vs darwin).
RUN npm install

COPY . .

RUN npm run build

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
