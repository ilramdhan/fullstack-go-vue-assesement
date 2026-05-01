FROM node:20-alpine AS builder

WORKDIR /src

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY openapi.yaml /openapi.yaml

COPY frontend/ ./

RUN sed -i "s|input: '../openapi.yaml'|input: '/openapi.yaml'|" openapi-ts.config.ts

ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm run build


FROM nginx:1.27-alpine

RUN apk add --no-cache wget

COPY --from=builder /src/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=3s --retries=3 \
  CMD wget -q -O- http://127.0.0.1/ || exit 1
