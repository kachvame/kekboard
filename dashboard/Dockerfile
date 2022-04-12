FROM node:alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN npm run build

FROM node:alpine AS app

WORKDIR /app
COPY --from=builder /app .
COPY . .

EXPOSE 3000
CMD ["node", "./build"]
