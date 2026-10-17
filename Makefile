.PHONY: ui build run dev clean

ui:
	cd ui && pnpm run build

build: ui
	go mod tidy
	go build -o local-cinema .

run: build
	./local-cinema

start: 
	./local-cinema

# Hot reload: Go server on :8765, Next dev server on :3000 proxying /api to it.
dev:
	@echo "Terminal 1: go run . serve --http=0.0.0.0:8765"
	@echo "Terminal 2: cd ui && npm run dev"

clean:
	rm -rf local-cinema pb_data ui/.next
