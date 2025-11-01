services: docker compose up db redis --remove-orphans
workerserver: go run src/ee/workerserver/main.go
hosting: go run src/ee/hosting/main.go
ui: cd ./src/ui && npm install && npm run dev -- --port 5400
