FROM mcr.microsoft.com/playwright:v1.48.0-jammy

WORKDIR /app
COPY package.json ./
RUN npm install
RUN npx playwright install chromium --with-deps

COPY server.mjs ./
EXPOSE 3001
CMD ["node", "server.mjs"]
