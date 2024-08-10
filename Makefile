run: build
	@./bin/app
build:
	@~/go/bin/templ generate && \
	 ./tailwindcss -i tailwind/css/app.css -o public/styles.css && \
		 go build -o bin/app .
css:
	./tailwindcss -i tailwind/css/app.css -o public/styles.css --watch
templ:
	~/go/bin/templ generate --watch --proxy=http://localhost:8000
build-js:
	@curl -sLo public/scripts/htmx.min.js https://cdn.jsdelivr.net/npm/htmx.org/dist/htmx.min.js && \
	curl -sLo public/scripts/alpine.js https://cdn.jsdelivr.net/npm/alpinejs/dist/cdn.min.js && \
	curl -sLo public/scripts/jquery.min.js https://cdn.jsdelivr.net/npm/jquery/dist/jquery.min.js && \
	curl -sLo public/scripts/jquery.waypoints.min.js https://cdn.jsdelivr.net/npm/waypoints/lib/jquery.waypoints.min.js
tailwind:
	@curl -sLo tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
	chmod +x tailwindcss