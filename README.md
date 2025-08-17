## m-data-storage

Сервер на TypeScript/Node для хранения исторических OHLCV (1m) в PostgreSQL/TimescaleDB.

### Быстрый старт (dev)

1. Установить зависимости: `yarn`
2. Скопировать `.env.example` в `.env` и заполнить `API_TOKEN`/`DATABASE_URL`
3. Запустить в dev-режиме: `yarn dev`
4. Сборка: `yarn build`; запуск собранного: `yarn start`

### Скрипты

-   `yarn dev`: запуск `ts-node src/main.ts`
-   `yarn build`: компиляция в `dist`
-   `yarn start`: запуск `node dist/main.js`
-   `yarn lint`: проверка ESLint
-   `yarn format`: форматирование Prettier
