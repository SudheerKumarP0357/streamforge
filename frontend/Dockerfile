# syntax=docker/dockerfile:1

ARG NODE_VERSION=24.11.0

FROM node:${NODE_VERSION}-alpine AS base

WORKDIR /app

FROM base AS build

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

FROM base AS runner

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup --system --gid 1001 nodejs \
    && adduser  --system --uid 1001 nextjs

WORKDIR /app

COPY --from=build /app/build/standalone /app/
COPY --from=build /app/build/static/ /app/build/static/

EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

ENTRYPOINT [ "node" ]
CMD [ "/app/server.js" ]