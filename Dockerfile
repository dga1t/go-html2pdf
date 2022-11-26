FROM surnet/alpine-wkhtmltopdf:3.9-0.12.5-full as wkhtmltopdf
FROM golang:alpine

# Install dependencies for wkhtmltopdf
RUN apk add --no-cache \
  libstdc++ \
  libx11 \
  libxrender \
  libxext \
  libssl1.1 \
  ca-certificates \
  fontconfig \
  freetype \
  ttf-dejavu \
  ttf-droid \
  ttf-freefont \
  ttf-liberation \
  && apk add --no-cache --virtual .build-deps \
  msttcorefonts-installer \
  \
  # Install microsoft fonts
  && update-ms-fonts \
  && fc-cache -f \
  \
  # Clean up when done
  && rm -rf /tmp/* \
  && apk del .build-deps

COPY --from=wkhtmltopdf /bin/wkhtmltopdf /bin/wkhtmltopdf

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o main main.go
EXPOSE 3333
ENTRYPOINT [ "/app/main" ]