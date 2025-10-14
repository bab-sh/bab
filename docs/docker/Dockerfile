FROM node:20-alpine AS builder
WORKDIR /app
COPY ../package.json ../yarn.lock ./
RUN yarn install --frozen-lockfile
COPY ../ .
RUN yarn docs:build

FROM nginx:1.27-alpine
COPY --from=builder /app/.vitepress/dist /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
