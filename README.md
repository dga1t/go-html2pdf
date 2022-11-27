# GO-HTML2PDF

HTML2PDF is a service written in golang, for converting HTML files into PDF. 

## Installation

- Install the [wkhtmltopdf](https://wkhtmltopdf.org/downloads.html) (only for local usage - skip if running in docker)
- Clone the repo:
```
git clone https://github.com/dga1t/go-html2pdf
cd go-html2pdf
```

## Usage

- Run locally:
```
go mod download
go build
./html2pdf
```
- Run in Docker:
```
docker build -t html2pdf .
docker run -p 3333:3333 -ti html2pdf
```
- To test the running app:
```
curl -i -X POST -F "file=@/***/testArchive.zip" http://localhost:3333/upload
```

## License

[MIT](https://choosealicense.com/licenses/mit/)