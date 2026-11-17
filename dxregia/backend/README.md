# Backend Go

La diagnostica controlla il listener DXSpider configurato all'avvio e mostra l'esito nel browser. Usa Go 1.22 o successivo e Node.js 22.12 o successivo con npm.

Dalla radice del repository:

~~~sh
cd dxregia/web
npm ci
npm run build
cd ../backend
go run ./cmd/dxregia -telnet 127.0.0.1:7300
~~~

Apri http://127.0.0.1:8080. Usa -port per cambiare la porta HTTP e -telnet HOST:PORTA per impostare il listener effettivo. L'anteprima HTTP ascolta solo sul loopback. Avvia DXSpider separatamente.
